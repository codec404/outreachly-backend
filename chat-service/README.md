# chat-service

The primary backend service for Outreachly. Handles authentication, user management, campaign orchestration, and email templates.

---

## Overview

| Concern       | Detail                                                    |
|---------------|-----------------------------------------------------------|
| Language      | Go 1.25                                                   |
| HTTP router   | [chi v5](https://github.com/go-chi/chi)                   |
| Database      | PostgreSQL 17                                             |
| DB driver     | pgx/v5 (queries) · lib/pq (migrations)                   |
| Auth          | JWT HS256 access tokens + stateful refresh tokens + Google OAuth 2.0 |
| Migrations    | golang-migrate (embedded in binary, runs on startup)      |
| Live reload   | [Air](https://github.com/air-verse/air)                   |
| Logging       | Uber Zap (JSON in prod · coloured console in dev)         |
| Run mode      | `RUN_MODE` env var — `server` (HTTP) or a worker name     |

---

## Architecture

The single binary behaves differently depending on the `RUN_MODE` environment variable:

| `RUN_MODE`           | What runs                                        |
|----------------------|--------------------------------------------------|
| `server` (or unset)  | HTTP server + background token-cleanup goroutine |
| `bulk-upload-worker` | Bulk-upload worker only — no HTTP server         |

```
main.go
  └── app.Init()              — load config, run migrations, seed super admin, open DB pool
  │
  ├── RUN_MODE=server (default)
  │     └── app.StartTokenCleanup()  — background goroutine, periodic cleanup
  │     └── app.StartServer()        — chi router, listen + graceful shutdown
  │
  └── RUN_MODE=<worker-name>
        └── app.RunWorker()          — looks up worker in registry, starts it, blocks on ctx
```

**Request path (server mode)**
```
HTTP → middleware (RequestID → TraceID → Authenticate → RequireRole)
     → router      (chi)
     → controller  (decode JSON · call service · render response)
     → service     (business logic · JWT/bcrypt · token rotation)
     → repository  (SQL via pgx)
     → PostgreSQL
```

**Package layout**

```
chat-service/
├── main.go
├── app/               # boot: config, migrations, seed, DB pool, HTTP server, worker registry
├── worker/            # background worker implementations (one file per worker)
├── router/            # route registration + dependency wiring
├── controller/        # HTTP handlers (one sub-package per domain)
│   ├── auth/
│   ├── admin/
│   ├── user/
│   ├── campaign/
│   ├── template/
│   └── health/
├── service/           # business logic interfaces + implementations
│   └── auth/
├── repository/        # data access interfaces + postgres implementations
│   ├── user/
│   ├── token/
│   └── oauth/
├── model/             # plain Go structs (no ORM tags)
├── dto/               # request / response shapes
│   └── auth/
├── middleware/        # Authenticate, RequireRole, context helpers
├── db/
│   └── migrations/    # embedded SQL files (*.up.sql / *.down.sql)
├── pkg/
│   ├── conc/          # SafeGo / SafeTry — panic-recovering goroutine helpers
│   ├── logger/        # Zap wrapper with trace ID support
│   ├── render/        # JSON response writer (sets os-trace-id header)
│   ├── errorhandler/  # structured error responses
│   └── external_error/# ExternalError — safe to expose to clients
└── configs/
    └── config.yml     # static server config (host, port)
```

---

## Database schema

| Table              | Purpose                                                                 |
|--------------------|-------------------------------------------------------------------------|
| `users`            | Accounts with soft delete, `is_blocked`, and `avatar_url`               |
| `oauth_providers`  | Linked OAuth identities (provider + provider_user_id → local user)      |
| `roles`            | `user`, `admin`, `super_admin`                                          |
| `permissions`      | Fine-grained permission strings                                         |
| `user_roles`       | Many-to-many users ↔ roles                                             |
| `role_permissions` | Many-to-many roles ↔ permissions                                       |
| `refresh_tokens`   | Stateful refresh tokens (SHA-256 hash stored)                           |
| `templates`        | Email templates (body or S3 URL)                                        |
| `recruiters`       | Recruiter profiles linked to users                                      |
| `campaigns`        | Outreach campaigns with status lifecycle                                |
| `campaign_targets` | Individual targets within a campaign                                    |
| `email_events`     | Delivery/open/click events with JSONB metadata                          |

All tables have `created_at` and `updated_at` (`TIMESTAMPTZ`), maintained by a shared PL/pgSQL trigger. Migrations are embedded in the binary and applied automatically at startup.

---

## Authentication

### Email / password

| Flow     | Endpoint                     | Description                            |
|----------|------------------------------|----------------------------------------|
| Register | `POST /api/v1/auth/register` | Creates account with `user` role       |
| Login    | `POST /api/v1/auth/login`    | Returns access + refresh token pair    |
| Refresh  | `POST /api/v1/auth/refresh`  | Rotates refresh token, issues new pair |
| Logout   | `POST /api/v1/auth/logout`   | Revokes the supplied refresh token     |

### Google OAuth 2.0

| Flow     | Endpoint                            | Description                                                  |
|----------|-------------------------------------|--------------------------------------------------------------|
| Initiate | `GET /api/v1/auth/google`           | Redirects browser to Google consent page                     |
| Callback | `GET /api/v1/auth/google/callback`  | Exchanges code, logs in or creates user, returns token pair  |

**How it works:**
1. Browser hits `GET /auth/google` → server redirects to Google with a CSRF state cookie.
2. User consents → Google redirects to `/auth/google/callback?code=...&state=...`.
3. Server validates the state cookie, exchanges the code for a Google access token, fetches user info, then:
   - If the Google account is already linked → logs in that user.
   - If a local account exists with the same email → links the Google identity to it.
   - Otherwise → creates a new `user`-role account and links it.
4. Returns the same `{ access_token, refresh_token, expires_in }` response as email/password login.

OAuth identities are stored in the `oauth_providers` table. A user's `avatar_url` is populated from Google only if no avatar is already set — existing avatars are never overwritten by OAuth logins.

**Access token** — HS256 JWT, 15 min TTL by default. Claims: `sub` (user ID), `email`, `roles`.

**Refresh token** — 32-byte cryptographically random value returned to the client. Only its SHA-256 hash is persisted in the database. Rotated on every use.

Protected routes require:
```
Authorization: Bearer <access_token>
```

---

## API endpoints

All routes are prefixed with `/api/v1`.

**Public**
```
GET  /health/ping

POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout
GET  /auth/google
GET  /auth/google/callback
```

**Authenticated** (`Authorization: Bearer <token>` required)
```
GET  /users/me
PUT  /users/me

GET    /templates
POST   /templates
GET    /templates/{templateID}
PUT    /templates/{templateID}
DELETE /templates/{templateID}

GET    /campaigns
POST   /campaigns
GET    /campaigns/{campaignID}
PUT    /campaigns/{campaignID}
DELETE /campaigns/{campaignID}
POST   /campaigns/{campaignID}/start
POST   /campaigns/{campaignID}/schedule
GET    /campaigns/{campaignID}/targets
POST   /campaigns/{campaignID}/targets
```

**Admin** (`admin` or `super_admin` role required)
```
GET    /admin/users
GET    /admin/users/{userID}
PUT    /admin/users/{userID}/block
PUT    /admin/users/{userID}/unblock
DELETE /admin/users/{userID}
```

**Super admin only**
```
PUT /admin/users/{userID}/role
```

---

## Response format

Every response includes an `os-trace-id` header for distributed tracing.

**Success**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "a3f9...",
  "expires_in": 900
}
```

**Error**
```json
{
  "code": 401,
  "message": "invalid credentials"
}
```

---

## Environment variables

Copy `local.env.example` to `local.env` and fill in the values.

| Variable                    | Required | Default  | Description                                                        |
|-----------------------------|----------|----------|--------------------------------------------------------------------|
| `DB_HOST`                   | yes      | —        | Postgres host                                                      |
| `DB_PORT`                   | yes      | —        | Postgres port                                                      |
| `DB_NAME`                   | yes      | —        | Database name                                                      |
| `DB_USER`                   | yes      | —        | Database user                                                      |
| `DB_PASSWORD`               | yes      | —        | Database password                                                  |
| `JWT_SECRET`                | yes      | —        | HS256 signing key (min 32 chars)                                   |
| `JWT_ACCESS_EXPIRY_MINUTES` | no       | `15`     | Access token TTL in minutes                                        |
| `JWT_REFRESH_EXPIRY_DAYS`   | no       | `7`      | Refresh token TTL in days                                          |
| `SUPER_ADMIN_NAME`          | yes      | —        | Default super admin display name                                   |
| `SUPER_ADMIN_EMAIL`         | yes      | —        | Default super admin email                                          |
| `SUPER_ADMIN_PASSWORD`      | yes      | —        | Default super admin password (min 8)                               |
| `GOOGLE_CLIENT_ID`          | no       | —        | OAuth 2.0 client ID from Google Cloud Console                      |
| `GOOGLE_CLIENT_SECRET`      | no       | —        | OAuth 2.0 client secret from Google Cloud Console                  |
| `GOOGLE_REDIRECT_URL`       | no       | —        | Must exactly match a URI registered in Google Cloud Console        |
| `APP_ENV`                   | no       | —        | Set to `production` to enable prod mode                            |
| `RUN_MODE`                  | no       | `server` | `server` starts the HTTP server; a worker name starts that worker  |

---

## Running locally

**Option A — full Docker (recommended)**
```bash
cp local.env.example local.env   # fill in values
make dev                          # starts postgres + pgadmin + chat-service + bulk-upload-worker
```

This starts four containers, each with the correct `RUN_MODE` already set:

| Container            | `RUN_MODE`           | What it does                  |
|----------------------|----------------------|-------------------------------|
| `postgres`           | —                    | Database                      |
| `pgadmin`            | —                    | DB admin UI                   |
| `chat-service`       | `server`             | HTTP API on `:8080`           |
| `bulk-upload-worker` | `bulk-upload-worker` | Bulk-upload worker (polling)  |

**Option B — infra in Docker, app on host**
```bash
cp local.env.example local.env
make infra-up          # starts only postgres + pgadmin
RUN_MODE=server go run .                    # run the HTTP server
RUN_MODE=bulk-upload-worker go run .        # run the worker (separate terminal)
```

The API is available at `http://localhost:8080`.

**pgAdmin** (dev only) → `http://localhost:5050`
Login with the credentials from `local.env`. The postgres connection is pre-configured.

For all available `make` targets see [COMMANDS.md](COMMANDS.md).

---

## Running in production

```bash
make prod
```

The prod image is a minimal Alpine binary (~10 MB). It reads all config from environment variables — no `local.env` is used in production.

In production (ECS), the same image is deployed as two separate services with different `RUN_MODE` values in the task definition environment:

| ECS Service          | `RUN_MODE`           |
|----------------------|----------------------|
| `chat-service`       | `server`             |
| `bulk-upload-worker` | `bulk-upload-worker` |

Setting an unrecognised `RUN_MODE` value causes the process to fail immediately at startup with a clear error listing valid values — misconfigured task definitions surface at deploy time, not silently at runtime.

---

## Workers

Workers are long-running background processes that consume jobs from a queue (SQS in production). They run as separate ECS services using the same Docker image as the HTTP server, differentiated only by `RUN_MODE`.

**How it works**

```
chat-service  ──push──▶  SQS queue  ◀──poll──  bulk-upload-worker
(HTTP server)                                   (ECS service, always running)
```

The server pushes a message and returns immediately (202 Accepted). The worker polls the queue in a loop, processes each message, then deletes it. SQS ensures each message is delivered to exactly one worker replica.

**Worker registry** — `app/workers.go`

```go
var workerFuncs = map[string]func(context.Context){
    BulkUploadWorker: worker.StartBulkUploadWorker,
    // add new workers here
}
```

The map key is the exact `RUN_MODE` value used in docker-compose / ECS task definitions. An unknown key causes the process to exit immediately at startup.

**Adding a new worker**

1. Create `worker/<name>.go` with a `StartXxxWorker(ctx context.Context)` function.
2. Add one line to `workerFuncs` in `app/workers.go`.
3. Add a new docker-compose service (dev) and ECS task definition (prod) with `RUN_MODE: <name>`.

**Goroutine safety** — all workers use `pkg/conc.SafeGo` for panic recovery and `SafeTry` per iteration so a single failing job does not kill the loop.

---

## Development notes

- **Migrations** run automatically at startup via embedded SQL files (`go:embed`). No CLI dependency required.
- **Super admin** is seeded on first boot. The seed is idempotent — safe to restart.
- **Google OAuth** — `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `GOOGLE_REDIRECT_URL` are optional at startup. If unset, the `/auth/google` routes return an error at call time. Register credentials at [console.cloud.google.com/apis/credentials](https://console.cloud.google.com/apis/credentials).
- **OAuth avatar policy** — `avatar_url` on the `users` table is populated from Google only if the field is currently empty. Manually set avatars are never overwritten.
- **Logs** are structured JSON in production, coloured console in dev. Every log line carries `trace_id` when available.
- **Panic recovery** — HTTP panics are caught by `jsonRecoverer`, logged with the trace ID, and returned as `500 Internal Server Error`. Worker panics are caught by `SafeGo` and logged; the worker goroutine exits but the process stays alive.
- **Graceful shutdown** — `SIGINT`/`SIGTERM` cancels the root context. The HTTP server drains in-flight requests; workers unblock from their poll loop and exit cleanly.
