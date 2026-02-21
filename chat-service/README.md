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
| Auth          | JWT HS256 access tokens + stateful refresh tokens         |
| Migrations    | golang-migrate (embedded in binary, runs on startup)      |
| Live reload   | [Air](https://github.com/air-verse/air)                   |
| Logging       | Uber Zap (JSON in prod · coloured console in dev)         |

---

## Architecture

```
main.go
  └── app.Init()         — load config, run migrations, seed super admin, open DB pool
  └── app.NewServer()    — build chi router, register all routes
  └── app.StartServer()  — listen + graceful shutdown

Request path:
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
├── app/               # boot: config, migrations, seed, DB pool, HTTP server
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
│   └── token/
├── model/             # plain Go structs (no ORM tags)
├── dto/               # request / response shapes
│   └── auth/
├── middleware/        # Authenticate, RequireRole, context helpers
├── migrations/        # embedded SQL files (*.up.sql / *.down.sql)
├── pkg/
│   ├── logger/        # Zap wrapper with trace ID support
│   ├── render/        # JSON response writer (sets os-trace-id header)
│   ├── errorhandler/  # structured error responses
│   └── external_error/# ExternalError — safe to expose to clients
└── configs/
    └── config.yml     # static server config (host, port)
```

---

## Database schema

| Table              | Purpose                                        |
|--------------------|------------------------------------------------|
| `users`            | Accounts with soft delete and `is_blocked`     |
| `roles`            | `user`, `admin`, `super_admin`                 |
| `permissions`      | Fine-grained permission strings                |
| `user_roles`       | Many-to-many users ↔ roles                    |
| `role_permissions` | Many-to-many roles ↔ permissions              |
| `refresh_tokens`   | Stateful refresh tokens (SHA-256 hash stored)  |
| `templates`        | Email templates (body or S3 URL)               |
| `recruiters`       | Recruiter profiles linked to users             |
| `campaigns`        | Outreach campaigns with status lifecycle       |
| `campaign_targets` | Individual targets within a campaign           |
| `email_events`     | Delivery/open/click events with JSONB metadata |

All tables have `created_at` and `updated_at` (`TIMESTAMPTZ`), maintained by a shared PL/pgSQL trigger. Migrations are embedded in the binary and applied automatically at startup.

---

## Authentication

| Flow     | Endpoint                     | Description                            |
|----------|------------------------------|----------------------------------------|
| Register | `POST /api/v1/auth/register` | Creates account with `user` role       |
| Login    | `POST /api/v1/auth/login`    | Returns access + refresh token pair    |
| Refresh  | `POST /api/v1/auth/refresh`  | Rotates refresh token, issues new pair |
| Logout   | `POST /api/v1/auth/logout`   | Revokes the supplied refresh token     |

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

| Variable                    | Required | Default | Description                             |
|-----------------------------|----------|---------|-----------------------------------------|
| `DB_HOST`                   | yes      | —       | Postgres host                           |
| `DB_PORT`                   | yes      | —       | Postgres port                           |
| `DB_NAME`                   | yes      | —       | Database name                           |
| `DB_USER`                   | yes      | —       | Database user                           |
| `DB_PASSWORD`               | yes      | —       | Database password                       |
| `JWT_SECRET`                | yes      | —       | HS256 signing key (min 32 chars)        |
| `JWT_ACCESS_EXPIRY_MINUTES` | no       | `15`    | Access token TTL in minutes             |
| `JWT_REFRESH_EXPIRY_DAYS`   | no       | `7`     | Refresh token TTL in days               |
| `SUPER_ADMIN_NAME`          | yes      | —       | Default super admin display name        |
| `SUPER_ADMIN_EMAIL`         | yes      | —       | Default super admin email               |
| `SUPER_ADMIN_PASSWORD`      | yes      | —       | Default super admin password (min 8)    |
| `APP_ENV`                   | no       | —       | Set to `production` to enable prod mode |

---

## Running locally

**Option A — full Docker (recommended)**
```bash
cp local.env.example local.env   # fill in values
make dev                          # starts postgres + pgadmin + chat-service with hot reload
```

**Option B — infra in Docker, app on host**
```bash
cp local.env.example local.env
make infra-up   # starts only postgres + pgadmin
go run .        # run the app directly (reads local.env automatically)
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

---

## Development notes

- **Migrations** run automatically at startup via embedded SQL files (`go:embed`). No CLI dependency required.
- **Super admin** is seeded on first boot. The seed is idempotent — safe to restart.
- **Logs** are structured JSON in production, coloured console in dev. Every log line carries `trace_id` when available.
- **Panic recovery** — all panics are caught by `jsonRecoverer`, logged with the trace ID, and returned as `500 Internal Server Error`.
