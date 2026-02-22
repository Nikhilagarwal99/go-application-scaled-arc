.PHONY: run dev build migrate-up migrate-down \
        up down up-infra up-app logs ps clean tidy

# ── Local (no Docker for app) ─────────────────────────────────────────────────

## run: run the server locally (needs postgres + redis running)
run:
	go run ./cmd/server/...

## dev: run with Air hot-reload locally (needs Air installed)
dev:
	air

## build: compile both binaries into bin/
build:
	go build -o bin/server  ./cmd/server/...
	go build -o bin/migrate ./cmd/migrate/...

## tidy: clean up go modules
tidy:
	go mod tidy

# ── Migrations ────────────────────────────────────────────────────────────────

## migrate-up: run all pending migrations
## migrate-up: run all pending migrations (uses Docker-mapped ports)
migrate-up:
	DB_HOST=localhost DB_PORT=7000 REDIS_ADDR=localhost:7001 go run ./cmd/migrate/... up

## migrate-down: roll back all migrations
migrate-down:
	DB_HOST=localhost DB_PORT=7000 REDIS_ADDR=localhost:7001 go run ./cmd/migrate/... down
	
# ── Docker ────────────────────────────────────────────────────────────────────

## up: start everything (postgres, redis, redis-commander, app)
up:
	docker-compose up -d

## up-infra: start only postgres + redis + redis-commander (run app locally)
up-infra:
	docker-compose up -d postgres redis redis-commander

## up-app: start only the app container
up-app:
	docker-compose up -d app

## down: stop and remove all containers
down:
	docker-compose down

## clean: stop containers and wipe volumes (WARNING: deletes all DB data)
clean:
	docker-compose down -v

## logs: tail logs from all containers
logs:
	docker-compose logs -f

## logs-app: tail logs from app container only
logs-app:
	docker-compose logs -f app

## ps: show running containers
ps:
	docker-compose ps

# ── Helpers ───────────────────────────────────────────────────────────────────

## install-air: install Air hot-reload tool
install-air:
	go install github.com/air-verse/air@latest