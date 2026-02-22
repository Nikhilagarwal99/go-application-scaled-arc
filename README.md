# Go Production Backend — Gin + GORM + PostgreSQL + Redis

A production-ready REST API with authentication, email verification (OTP via Mailjet), and Redis-backed OTP storage.

---

## Project Structure

```
.
├── cmd/
│   └── main.go              # Entrypoint — config, DB/Redis, server, graceful shutdown
├── internal/
│   ├── cache/               # Redis client (OTP storage, session cache)
│   ├── config/              # Environment config loader
│   ├── database/            # PostgreSQL connection & auto-migration
│   ├── handlers/            # HTTP handlers (auth, user profile)
│   ├── middleware/          # JWT auth, CORS, logging
│   ├── models/              # GORM models (User, etc.)
│   ├── repository/          # Data access (user_repository, otp_repository)
│   ├── server/              # Router & dependency wiring
│   ├── services/            # Business logic (auth, email verification)
│   └── utils/               # JWT, OTP, Mailjet email
├── pkg/
│   └── response/            # Standard JSON response helpers
├── .env.example
├── .air.toml                # Hot-reload config (Air)
├── docker-compose.yml       # Postgres, Redis, Redis Commander, app
├── Dockerfile               # Production build
├── Dockerfile.dev           # Dev image with Air
└── go.mod
```

---

## Quick Start

### 1. Clone & install dependencies

```bash
go mod tidy
```

### 2. Copy and configure environment

```bash
cp .env.example .env
```

Edit `.env`: set `DB_PASSWORD`, and optionally Mailjet keys for email verification. When using Docker Compose for Postgres/Redis, use:

- `DB_HOST=localhost` (or `postgres` when running app inside Docker)
- `DB_PORT=7000` (host port mapped from Postgres in docker-compose)
- `REDIS_ADDR=localhost:7001` (or `redis:6379` when app is in Docker)

### 3. Start dependencies (Postgres + Redis) with Docker

```bash
docker-compose up -d postgres redis
```

Optional: start Redis Commander for debugging Redis:

```bash
docker-compose up -d redis-commander
# UI at http://localhost:7002
```

### 4. Run the server

**Locally (with Postgres/Redis in Docker):**

```bash
go run ./cmd/main.go
```

Server runs on `SERVER_PORT` (default `7003` from `.env.example`). Health check: `GET http://localhost:7003/health`

**Full stack in Docker (app + Postgres + Redis):**

```bash
docker-compose up -d
```

App is exposed on port `7003`; ensure `.env` has `DB_HOST=postgres` and `REDIS_ADDR=redis:6379` (docker-compose overrides these when running the app service).

**Development with hot-reload (Air):**

```bash
air
```

Uses `.air.toml` to rebuild and restart on Go file changes.

---

## API Reference

Responses use a consistent envelope:

- Success: `{ "success": true, "message": "...", "data": { ... } }`
- Error: `{ "success": false, "error": "..." }` (with appropriate HTTP status)

### Health

| Method | Path     | Description   |
|--------|----------|---------------|
| GET    | `/health` | Service health |

### Auth (public)

| Method | Path                          | Body                           | Description              |
|--------|--------------------------------|--------------------------------|--------------------------|
| POST   | `/api/v1/auth/signup`          | `{ "name", "email", "password" }` | Create account           |
| POST   | `/api/v1/auth/login`           | `{ "email", "password" }`      | Login → JWT token        |
| POST   | `/api/v1/auth/send-verify-email-otp` | `{ "email" }`            | Send OTP to email (Mailjet) |
| POST   | `/api/v1/auth/verify-email`    | `{ "email", "otp" }`           | Verify email with OTP    |

### Users (protected)

Require header: `Authorization: Bearer <token>`.

| Method | Path               | Body        | Description         |
|--------|--------------------|-------------|---------------------|
| GET    | `/api/v1/users/`   | —           | Get own profile     |
| PUT    | `/api/v1/users/`   | `{ "name" }`| Update profile      |
| DELETE | `/api/v1/users/`   | —           | Soft-delete account |

---

## Environment Variables

| Variable              | Description                    | Default / Example     |
|-----------------------|--------------------------------|------------------------|
| `APP_ENV`             | `development` / `production`   | `development`         |
| `SERVER_PORT`         | HTTP server port               | `7003`                 |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` | PostgreSQL | — |
| `JWT_SECRET`          | Secret for signing JWTs        | —                      |
| `JWT_EXPIRY_HOURS`    | Token expiry in hours          | `24`                   |
| `REDIS_ADDR`          | Redis address                  | `localhost:6379`       |
| `REDIS_PASSWORD`      | Redis password (if any)        | —                      |
| `REDIS_DB`            | Redis DB index                 | `0`                    |
| `MAILJET_API_KEY`, `MAILJET_API_SECRET` | Mailjet API credentials | — |
| `MAILJET_SENDER_EMAIL`, `MAILJET_SENDER_NAME` | Sender for verification emails | — |

---

## Design Decisions

- **Repository pattern** — DB and Redis access behind interfaces for testability.
- **Service layer** — business logic in services; handlers stay thin.
- **UUIDs** as primary keys — no sequential ID exposure.
- **Soft deletes** — GORM `DeletedAt`; user records are not hard-deleted.
- **bcrypt** for passwords — password field excluded from JSON.
- **Email verification** — OTP stored in Redis, sent via Mailjet; verified flag on user.
- **Graceful shutdown** — SIGINT/SIGTERM handled; in-flight requests complete before exit.
- **Connection tuning** — PostgreSQL pool and timeouts configured for moderate load.
