# ── Shell ─────────────────────────────────────────────────────────────────────
SHELL        := bash
.SHELLFLAGS  := -euo pipefail -c

# ── Config ────────────────────────────────────────────────────────────────────
# -include: silently skips if local.env is absent (CI uses real env vars).
# export:   makes every Make variable available to child shell commands.
-include local.env
export

MIGRATIONS_DIR := ./db/migrations
DB_DSN         := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# ── Default ───────────────────────────────────────────────────────────────────
.DEFAULT_GOAL := help

.PHONY: help \
        dev prod build down down-v ps restart logs \
        infra-up infra-down clean \
        tidy fmt vet lint test check \
        migrate-down migrate-status migrate-create migrate-force \
        shell-db

# ── Help ──────────────────────────────────────────────────────────────────────
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Docker ────────────────────────────────────────────────────────────────────
dev: ## Start all services in dev mode (hot reload via air)
	docker compose up --build -d

prod: ## Start all services in prod mode
	docker compose -f docker-compose.yml up --build -d

build: ## Build the prod image without starting containers
	docker compose -f docker-compose.yml build

down: ## Stop and remove containers (volumes are preserved)
	docker compose down

down-v: ## Stop and remove containers AND volumes  ⚠ destructive
	@echo "WARNING: this will delete all postgres data. Press Ctrl-C to abort."; sleep 3
	docker compose down --volumes

infra-up: ## Start only infrastructure services (postgres, pgadmin) — run the app locally
	docker compose up -d postgres pgadmin

infra-down: ## Stop only infrastructure services
	docker compose stop postgres pgadmin

clean: ## Remove containers, volumes, images, and build cache  ⚠ destructive
	@echo "WARNING: this will delete all containers, volumes, and images. Press Ctrl-C to abort."; sleep 3
	docker compose down --volumes --rmi local
	docker builder prune -f

ps: ## Show status of running containers
	docker compose ps

restart: ## Restart all services
	docker compose restart

logs: ## Tail logs for all services
	docker compose logs -f

logs-%: ## Tail logs for one service  (e.g. make logs-postgres)
	docker compose logs -f $*

shell-db: ## Open a psql shell inside the postgres container
	docker compose exec postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

# ── Go ────────────────────────────────────────────────────────────────────────
tidy: ## Tidy go.mod and go.sum
	go mod tidy

fmt: ## Format all Go source files
	gofmt -w .

vet: ## Run go vet across all packages
	go vet ./...

lint: ## Run golangci-lint  (install: brew install golangci-lint)
	golangci-lint run ./...

test: ## Run all tests with race detector
	go test -race -count=1 ./...

check: fmt vet test ## Run fmt + vet + test  (pre-push gate)

# ── Migrations ────────────────────────────────────────────────────────────────
# migrate-up is intentionally absent: migrations run automatically on app startup
# via go:embed + golang-migrate library (no CLI dependency).
#
# The targets below are for manual ad-hoc operations only.
# They require the golang-migrate CLI:
#   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-down: ## Roll back the last applied migration  (manual)
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down 1

migrate-status: ## Print the current migration version
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" version

migrate-create: ## Scaffold a new migration pair  (usage: make migrate-create NAME=add_index_to_users)
	@[ -n "$(NAME)" ] || { echo "ERROR: NAME is required — make migrate-create NAME=<description>"; exit 1; }
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

migrate-force: ## Force-set migration version to fix a dirty state  (usage: make migrate-force VERSION=7)
	@[ -n "$(VERSION)" ] || { echo "ERROR: VERSION is required — make migrate-force VERSION=<n>"; exit 1; }
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" force $(VERSION)
