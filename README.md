# Go Production Backend

A production-ready REST API built with Go, Gin, GORM, PostgreSQL, Redis, and Asynq background workers.

---

## Tech Stack

| Layer            | Technology                |
|------------------|---------------------------|
| Framework        | Gin                       |
| ORM              | GORM                      |
| Database         | PostgreSQL (Master/Slave) |
| Cache / OTP      | Redis                     |
| Auth             | JWT                       |
| Email            | Mailjet                   |
| Logging          | Zap                       |
| Migrations       | golang-migrate            |
| Background Jobs  | Asynq (Redis-backed)      |
| Hot Reload       | Air                       |
| Containerization | Docker + Docker Compose   |

---

## Project Structure
```
.
├── cmd/
│   ├── server/
│   │   ├── main.go               # Entrypoint — boots logger, DB, Redis, HTTP server
│   │   └── router.go             # Route registration + dependency wiring
│   ├── worker/
│   │   └── main.go               # Worker entrypoint — processes background tasks
│   └── migrate/
│       └── main.go               # Standalone migration binary
├── internal/
│   ├── cache/
│   │   └── redis.go              # Redis client — Set, Get, Delete, Ping
│   ├── config/
│   │   └── config.go             # Env config loader
│   ├── database/
│   │   ├── database.go           # PostgreSQL connect, migrate, ping
│   │   └── migrations/
│   │       ├── 000001_create_users.up.sql
│   │       └── 000001_create_users.down.sql
│   ├── handlers/
│   │   ├── auth_handler.go       # Auth HTTP handlers
│   │   └── health_handler.go     # Health check handler
│   ├── logger/
│   │   └── logger.go             # Zap logger — dev console + prod JSON
│   ├── middleware/
│   │   ├── auth.go               # JWT validation
│   │   ├── middleware.go         # CORS
│   │   ├── logger.go             # Request logger (zap)
│   │   ├── request_id.go         # Unique request ID per request
│   │   └── transaction.go        # Auto DB transaction — BEGIN/COMMIT/ROLLBACK
│   ├── models/
│   │   └── user.go               # User GORM model
│   ├── repository/
│   │   ├── user_repository.go    # User DB access — tx aware via getDB(ctx)
│   │   └── otp_repository.go     # OTP Redis access
│   ├── services/
│   │   └── auth_service.go       # Business logic — signup, login, OTP, verify
│   ├── tasks/
│   │   ├── client.go             # Enqueue jobs from services
│   │   └── email_task.go         # Task definitions + processors
│   └── utils/
│       ├── jwt.go                # JWT generate + validate
│       ├── otp.go                # Crypto random OTP generator
│       └── mailjet.go            # Mailjet email sender
├── pkg/
│   ├── errorType/
│   │   └── errors.go             # Centralized typed AppError
│   └── response/
│       └── response.go           # Standard JSON envelope + error mapping
├── .air.toml                     # Air hot-reload config
├── .env.example                  # Environment variable template
├── docker-compose.yml            # Full stack — postgres, redis, app, worker
├── Dockerfile                    # Production HTTP server build
├── Dockerfile.dev                # Development build with Air
└── Dockerfile.worker             # Production worker build
```

---

## Quick Start

### 1. Clone and install dependencies
```bash
git clone https://github.com/Nikhilagarwal99/go-application-scaled-arc.git
cd go-application-scaled-arc
go mod tidy
```

### 2. Configure environment
```bash
cp .env.example .env
# fill in DB_PASSWORD, JWT_SECRET, Mailjet keys
```

### 3. Start infrastructure
```bash
make up-infra    # starts postgres + redis + redis-commander
```

### 4. Run migrations
```bash
make migrate-up
```

### 5. Start server + worker
```bash
make dev         # HTTP server with hot-reload
make up-worker   # worker in Docker (separate terminal)
```

---

## Docker (Full Stack)
```bash
make up            # start everything including worker
make logs-app      # watch HTTP server logs
make logs-worker   # watch worker logs
make down          # stop everything
make clean         # stop + wipe volumes (WARNING: deletes all data)
```

---

## Ports

| Service         | Host Port | Container Port |
|-----------------|-----------|----------------|
| App             | 7003      | 7003           |
| PostgreSQL      | 7000      | 5432           |
| Redis           | 7001      | 6379           |
| Redis Commander | 7002      | 8081           |

---

## Makefile Commands

| Command              | Description                              |
|----------------------|------------------------------------------|
| `make run`           | Run server locally                       |
| `make dev`           | Run with Air hot-reload                  |
| `make build`         | Compile server + worker + migrate        |
| `make tidy`          | Clean up go modules                      |
| `make migrate-up`    | Run all pending migrations               |
| `make migrate-down`  | Roll back all migrations                 |
| `make up`            | Start all Docker services                |
| `make up-infra`      | Start postgres + redis only              |
| `make up-app`        | Start app container only                 |
| `make up-worker`     | Start worker container only              |
| `make down`          | Stop all containers                      |
| `make clean`         | Stop containers + wipe volumes           |
| `make logs`          | Tail all container logs                  |
| `make logs-app`      | Tail app container logs                  |
| `make logs-worker`   | Tail worker container logs               |
| `make ps`            | Show running containers                  |
| `make scale-worker N=3` | Scale worker to N instances          |
| `make install-air`   | Install Air hot-reload tool              |

---

## API Reference

All responses follow this envelope:
```json
// success
{ "success": true, "message": "...", "data": {} }

// error
{ "success": false, "code": "ERROR_CODE", "message": "..." }
```

### Health

| Method | Path      | Auth | Description                         |
|--------|-----------|------|-------------------------------------|
| GET    | `/health` | —    | Checks postgres + redis live status |

### Auth (Public)

| Method | Path                                 | Description           |
|--------|--------------------------------------|-----------------------|
| POST   | `/api/v1/auth/signup`                | Create account        |
| POST   | `/api/v1/auth/login`                 | Login → JWT token     |
| POST   | `/api/v1/auth/send-verify-email-otp` | Send OTP to email     |
| POST   | `/api/v1/auth/verify-email`          | Verify email with OTP |

### Users (Protected)

Require header: `Authorization: Bearer <token>`

| Method | Path             | Description         |
|--------|------------------|---------------------|
| GET    | `/api/v1/users/` | Get own profile     |
| PUT    | `/api/v1/users/` | Update name         |
| DELETE | `/api/v1/users/` | Soft-delete account |

---

## Background Jobs

Email sending is fully decoupled from HTTP requests using Asynq backed by Redis.

### How it works
```
User requests OTP
  ↓
Service saves OTP to Redis
  ↓
Service enqueues email:verify task  ← returns in <1ms
  ↓
User gets 200 immediately
  ↓                         (background)
Worker picks up task from Redis
  ↓
Calls Mailjet
  ↓
Fails? → retries 3x with exponential backoff
All retries fail? → moves to dead letter queue
```

### Task types

| Task            | Queue    | Trigger       | Retries |
|-----------------|----------|---------------|---------|
| `email:verify`  | critical | Request OTP   | 3       |
| `email:welcome` | default  | User signup   | 3       |

### Queue priorities
```
critical → processed 3x more than default
default  → processed when critical queue is clear
```

This ensures OTP emails (user is actively waiting) always take priority over welcome emails (background, no urgency).

### Scaling workers during traffic spikes
```bash
make scale-worker N=5   # spin up 5 workers to drain queue
make scale-worker N=1   # scale back down when queue clears
```

Workers scale independently from HTTP servers — no app restart needed.

---

## Environment Variables

| Variable               | Description                       | Example          |
|------------------------|-----------------------------------|------------------|
| `APP_ENV`              | `development` or `production`     | `development`    |
| `SERVER_PORT`          | HTTP port                         | `7003`           |
| `DB_HOST`              | Postgres host                     | `localhost`      |
| `DB_PORT`              | Postgres port (host mapped)       | `7000`           |
| `DB_USER`              | Postgres user                     | `postgres`       |
| `DB_PASSWORD`          | Postgres password                 | —                |
| `DB_NAME`              | Postgres database name            | `goapp_db`       |
| `DB_SSLMODE`           | SSL mode                          | `disable`        |
| `DB_SLAVE_HOST`        | Slave host (falls back to master) | `localhost`      |
| `DB_SLAVE_PORT`        | Slave port                        | `7000`           |
| `JWT_SECRET`           | JWT signing secret                | —                |
| `JWT_EXPIRY_HOURS`     | Token expiry in hours             | `24`             |
| `REDIS_ADDR`           | Redis address                     | `localhost:7001` |
| `REDIS_PASSWORD`       | Redis password                    | —                |
| `REDIS_DB`             | Redis DB index                    | `0`              |
| `MAILJET_API_KEY`      | Mailjet API key                   | —                |
| `MAILJET_API_SECRET`   | Mailjet secret                    | —                |
| `MAILJET_SENDER_EMAIL` | Sender email address              | —                |
| `MAILJET_SENDER_NAME`  | Sender display name               | —                |

---

## Architecture
```
Request
  ↓
gin.Recovery()          → catches panics
CORS()                  → sets headers
RequestID()             → assigns unique trace ID per request
RequestLogger()         → structured zap logging per request
Auth()                  → validates JWT (protected routes only)
Transaction()           → BEGIN tx (write routes only)
  ↓
Handler                 → validates request body (ShouldBindJSON)
  ↓
Service                 → business logic + typed AppErrors
  ↓
Repository              → DB via master/slave + transaction aware
Tasks Client            → enqueues background jobs to Redis
  ↓
Response                → standard envelope + automatic error mapping


Redis Queue (Asynq)
  ↓
Worker
  ↓
Task Processor          → email sending, retries, dead letter
```

---

## Middleware Per Route

| Route                               | Auth | Transaction |
|-------------------------------------|------|-------------|
| POST `/auth/signup`                 | —    | ✓           |
| POST `/auth/login`                  | —    | —           |
| POST `/auth/send-verify-email-otp`  | —    | —           |
| POST `/auth/verify-email`           | —    | ✓           |
| GET  `/users/`                      | ✓    | —           |
| PUT  `/users/`                      | ✓    | ✓           |
| DELETE `/users/`                    | ✓    | ✓           |

---

## Design Decisions

- **Repository pattern** — all DB/Redis access behind interfaces, easy to mock in tests
- **Service layer** — business logic isolated from HTTP concerns
- **Typed errors** — `AppError` carries HTTP status + error code, no string matching in handlers
- **Automatic transactions** — middleware owns BEGIN/COMMIT/ROLLBACK, services never touch it
- **Master/Slave splitting** — writes to master, reads to slave via GORM dbresolver
- **Request ID** — every log line carries a trace ID for end-to-end debugging
- **Background workers** — email sending fully decoupled from HTTP via Asynq + Redis
- **Worker scaling** — workers scale independently from HTTP servers
- **UUIDs** as primary keys — no sequential ID enumeration
- **Soft deletes** — records never hard deleted, `deleted_at` used
- **Graceful shutdown** — both HTTP server and worker drain in-flight work before exit
- **Connection pooling** — max open/idle conns + lifetime configured for production load