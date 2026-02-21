# Make commands — chat-service

Run `make help` inside `chat-service/` to print this list at any time.

All commands read `local.env` automatically (if it exists) so you never need to pass DB credentials by hand.

---

## Docker

### `make dev`
Start the full stack in **dev mode** with hot reload via Air.

Starts: `postgres` + `pgadmin` + `chat-service` (Air watches for file changes and rebuilds automatically).

```bash
make dev
```

> Migrations run automatically when the app starts. The super admin is seeded on first boot.

---

### `make prod`
Start the full stack in **production mode** using the multi-stage Alpine image.

```bash
make prod
```

Uses `docker-compose.yml` only (no override file).

---

### `make build`
Build the production Docker image without starting any containers.

```bash
make build
```

---

### `make infra-up`
Start **only** the infrastructure services (postgres + pgadmin). Useful when you want to run the Go app directly on your host with `go run .`.

```bash
make infra-up
go run .
```

---

### `make infra-down`
Stop the infrastructure services (postgres + pgadmin) without removing them.

```bash
make infra-down
```

---

### `make down`
Stop and remove all containers. Volumes (database data) are preserved.

```bash
make down
```

---

### `make down-v`
⚠ **Destructive.** Stop and remove all containers **and volumes** (all postgres data is wiped).

Prompts with a 3-second countdown before executing.

```bash
make down-v
```

---

### `make clean`
⚠ **Destructive.** Remove all containers, volumes, local images, and Docker build cache.

Prompts with a 3-second countdown before executing.

```bash
make clean
```

---

### `make ps`
Show the status of all running containers.

```bash
make ps
```

---

### `make restart`
Restart all running services.

```bash
make restart
```

---

### `make logs`
Tail logs from all services.

```bash
make logs
```

---

### `make logs-<service>`
Tail logs from a specific service.

```bash
make logs-postgres
make logs-chat-service
make logs-pgadmin
```

---

### `make shell-db`
Open a `psql` shell inside the running postgres container.

```bash
make shell-db
```

---

## Go

### `make tidy`
Tidy `go.mod` and `go.sum`.

```bash
make tidy
```

---

### `make fmt`
Format all Go source files with `gofmt`.

```bash
make fmt
```

---

### `make vet`
Run `go vet` across all packages.

```bash
make vet
```

---

### `make lint`
Run `golangci-lint` across all packages.

```bash
make lint
```

> Requires `golangci-lint`. Install with:
> ```bash
> brew install golangci-lint
> ```

---

### `make test`
Run all tests with the race detector enabled.

```bash
make test
```

---

### `make check`
Run `fmt` + `vet` + `test` in sequence. Use this as a pre-push gate.

```bash
make check
```

---

## Migrations

> **`migrate-up` does not exist** — migrations run automatically when the app boots.
> The targets below are for manual, ad-hoc operations only.

All migration targets require the `golang-migrate` CLI:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

### `make migrate-down`
Roll back the **last applied** migration by one step.

```bash
make migrate-down
```

---

### `make migrate-status`
Print the current migration version and whether the database is in a dirty state.

```bash
make migrate-status
```

---

### `make migrate-create NAME=<description>`
Scaffold a new numbered migration pair (`*.up.sql` + `*.down.sql`) in `./migrations/`.

```bash
make migrate-create NAME=add_index_to_email_events
# creates: migrations/000015_add_index_to_email_events.up.sql
#          migrations/000015_add_index_to_email_events.down.sql
```

---

### `make migrate-force VERSION=<n>`
Force-set the migration version to a specific number. Use this to fix a **dirty** migration state after a failed partial migration.

```bash
make migrate-force VERSION=14
```

> Only use this if you have manually cleaned up the half-applied migration in the database first.
