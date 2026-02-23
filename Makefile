.PHONY: run dev build migrate-up migrate-down \
        up down up-infra up-app up-worker logs ps clean tidy

# ── Local ─────────────────────────────────────────────────────────────────────

## run: run the HTTP server locally
run:
	go run ./cmd/server/...

## dev: run with Air hot-reload locally
dev:
	air

## build: compile server + worker + migrate binaries
build:
	go build -o bin/server  ./cmd/server/...
	go build -o bin/worker  ./cmd/worker/...
	go build -o bin/migrate ./cmd/migrate/...

## tidy: clean up go modules
tidy:
	go mod tidy

# ── Migrations ────────────────────────────────────────────────────────────────

## migrate-up: run all pending migrations (uses Docker-mapped ports)
migrate-up:
	DB_HOST=localhost DB_PORT=7000 REDIS_ADDR=localhost:7001 go run ./cmd/migrate/... up

## migrate-down: roll back all migrations
migrate-down:
	DB_HOST=localhost DB_PORT=7000 REDIS_ADDR=localhost:7001 go run ./cmd/migrate/... down

## migrate-step-down: roll back only the last migration
migrate-step-down:
	DB_HOST=localhost DB_PORT=7000 REDIS_ADDR=localhost:7001 go run ./cmd/migrate/... step-down

## migrate-version: show current migration version
migrate-version:
	DB_HOST=localhost DB_PORT=7000 REDIS_ADDR=localhost:7001 go run ./cmd/migrate/... version

# ── Docker ────────────────────────────────────────────────────────────────────

## up: start everything (postgres, redis, redis-commander, app, worker)
up:
	docker-compose up -d

## up-infra: start only postgres + redis + redis-commander
up-infra:
	docker-compose up -d postgres redis redis-commander

## up-app: start only the HTTP app container
up-app:
	docker-compose up -d app

## up-worker: start only the worker container
up-worker:
	docker-compose up -d worker

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

## logs-worker: tail logs from worker container only
logs-worker:
	docker-compose logs -f worker

## ps: show running containers
ps:
	docker-compose ps

## scale-worker: scale worker to N instances e.g. make scale-worker N=3
scale-worker:
	docker-compose up -d --scale worker=$(N) --no-recreate

# ── Helpers ───────────────────────────────────────────────────────────────────

## install-air: install Air hot-reload tool
install-air:
	go install github.com/air-verse/air@latest