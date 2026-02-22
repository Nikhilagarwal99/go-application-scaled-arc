# Go Production Backend

A production-ready REST API built with Go, Gin, GORM, PostgreSQL, and Redis.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL (Master/Slave) |
| Cache / OTP | Redis |
| Auth | JWT |
| Email | Mailjet |
| Logging | Zap |
| Migrations | golang-migrate |
| Hot Reload | Air |
| Containerization | Docker + Docker Compose |

---

## Project Structure
```
.
├── cmd/
│   ├── server/
│   │   ├── main.go          # Entrypoint — boots server, graceful shutdown
│   │   └── router.go        # Route registration & dependency wiring
│   └── migrate/
│       └── main.go          # Standalone migration binary
├── internal/
│   ├── cache/               # Redis client
│   ├── config/              # Environment config loader
│   ├── database/            # PostgreSQL connection, migrations
│   │   └── migrations/      # SQL migration files
│   ├── handlers/            # HTTP handlers — thin, delegate to services
│   ├── logger/              # Zap structured logger
│   ├── middleware/          # JWT auth, CORS, RequestID, Logger, Transaction
│   ├── models/              # GORM models
│   ├── repository/          # Data access layer — DB and Redis
│   ├── server/              # Router setup
│   ├── services/            # Business logic
│   └── utils/               # JWT, OTP, Mailjet
├── pkg/
│   ├── apperrors/           # Centralized typed errors
│   └── response/            # Standard JSON response envelope
├── .air.toml                # Air hot-reload config
├── .env.example             # Environment variable template
├── docker-compose.yml       # Full stack — app, postgres, redis
├── Dockerfile               # Production build
└── Dockerfile.dev           # Development build with Air
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
# edit .env and fill in your values
```

### 3. Start infrastructure
```bash
make up-infra    # starts postgres, redis, redis-commander
```

### 4. Run migrations
```bash
make migrate-up
```

### 5. Start the server
```bash
make dev         # with Air hot-reload
# or
make run         # without hot-reload
```

---

## Docker (Full Stack)
```bash
make up          # start everything
make logs-app    # watch app logs
make down        # stop everything
make clean       # stop + wipe volumes (WARNING: deletes all data)
```

---

## Makefile Commands

| Command | Description |
|---|---|
| `make run` | Run server locally |
| `make dev` | Run with Air hot-reload |
| `make build` | Compile server + migrate binaries |
| `make tidy` | Clean up go modules |
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Roll back all migrations |
| `make up` | Start all Docker services |
| `make up-infra` | Start postgres + redis only |
| `make up-app` | Start app container only |
| `make down` | Stop all containers |
| `make clean` | Stop containers + wipe volumes |
| `make logs` | Tail all container logs |
| `make logs-app` | Tail app container logs |
| `make ps` | Show running containers |
| `make install-air` | Install Air hot-reload tool |

---

## API Reference

All responses follow this envelope:
```json
// success
{ "success": true, "message": "...", "data": { } }

// error
{ "success": false, "code": "ERROR_CODE", "message": "..." }
```

### Health

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | — | Checks postgres + redis connectivity |

### Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/auth/signup` | — | Create account |
| POST | `/api/v1/auth/login` | — | Login → JWT |
| POST | `/api/v1/auth/send-verify-email-otp` | — | Send OTP to email |
| POST | `/api/v1/auth/verify-email` | — | Verify email with OTP |

### Users

Require header: `Authorization: Bearer <token>`

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/users/` | ✓ | Get own profile |
| PUT | `/api/v1/users/` | ✓ | Update name |
| DELETE | `/api/v1/users/` | ✓ | Soft-delete account |

---

## Environment Variables

| Variable | Description | Example |
|---|---|---|
| `APP_ENV` | `development` or `production` | `development` |
| `SERVER_PORT` | HTTP port | `7003` |
| `DB_HOST` | Postgres host | `localhost` |
| `DB_PORT` | Postgres port (host mapped) | `7000` |
| `DB_USER` | Postgres user | `postgres` |
| `DB_PASSWORD` | Postgres password | — |
| `DB_NAME` | Postgres database | `goapp_db` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `DB_SLAVE_HOST` | Slave host (falls back to master) | `localhost` |
| `DB_SLAVE_PORT` | Slave port | `7000` |
| `JWT_SECRET` | JWT signing secret | — |
| `JWT_EXPIRY_HOURS` | Token expiry | `24` |
| `REDIS_ADDR` | Redis address | `localhost:7001` |
| `REDIS_PASSWORD` | Redis password | — |
| `REDIS_DB` | Redis DB index | `0` |
| `MAILJET_API_KEY` | Mailjet API key | — |
| `MAILJET_API_SECRET` | Mailjet API secret | — |
| `MAILJET_SENDER_EMAIL` | Sender email | — |
| `MAILJET_SENDER_NAME` | Sender name | — |

---

## Architecture
```
Request
  ↓
gin.Recovery()          → catches panics
CORS()                  → sets headers
RequestID()             → assigns unique trace ID
RequestLogger()         → structured zap logging
Auth()                  → validates JWT (protected routes)
Transaction()           → BEGIN tx (write routes only)
  ↓
Handler                 → validates request body
  ↓
Service                 → business logic + typed AppErrors
  ↓
Repository              → DB (master/slave aware + tx aware)
  ↓
Response                → standard envelope + automatic error mapping
```

## Design Decisions

- **Repository pattern** — all DB/Redis access behind interfaces, easy to mock in tests
- **Service layer** — business logic isolated from HTTP concerns
- **Typed errors** — `AppError` carries HTTP status + app error code, no string matching
- **Automatic transactions** — middleware owns BEGIN/COMMIT/ROLLBACK, handlers never touch it
- **Master/Slave splitting** — writes go to master, reads go to slave via dbresolver
- **UUIDs** as primary keys — no sequential ID enumeration
- **Soft deletes** — records never hard deleted, `deleted_at` used
- **Request ID** — every log line carries a trace ID for end-to-end debugging
- **Graceful shutdown** — in-flight requests finish before process exits