# Outreachly — Backend

Outreachly is a recruitment outreach platform. This monorepo contains all backend microservices.

---

## Repository structure

```
outreachly-backend/
├── chat-service/      # Core API — auth, users, campaigns, templates
├── bloom-service/     # (planned)
├── mail-service/      # (planned)
├── worker-service/    # (planned)
└── LICENSE
```

Each service is a fully self-contained Go module with its own `go.mod`, `Dockerfile`, and `docker-compose` files.

---

## Services

| Service | Description | Stack |
|---------|-------------|-------|
| [chat-service](chat-service/README.md) | Core API — auth, users, campaigns, templates | Go · chi · PostgreSQL · JWT |
| bloom-service | _(planned)_ | — |
| mail-service | _(planned)_ | — |
| worker-service | _(planned)_ | — |

---

## Quick start

```bash
# 1 — Clone
git clone <repo-url>
cd outreachly-backend

# 2 — Set up environment
cp chat-service/local.env.example chat-service/local.env
# Edit local.env with your values

# 3 — Start (dev mode with hot reload)
cd chat-service
make dev
```

The API will be available at `http://localhost:8080`.

For full setup instructions, architecture, API reference, and environment variables see the **[chat-service README](chat-service/README.md)**.
