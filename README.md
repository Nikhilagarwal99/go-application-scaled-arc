# Go Production Backend — Gin + GORM + PostgreSQL

A clean, production-ready REST API boilerplate with authentication.

---

## Project Structure

```
.
├── cmd/
│   └── server/
│       ├── main.go        # Entrypoint — boots server, graceful shutdown
│       └── router.go      # Route registration & dependency wiring
├── internal/
│   ├── config/            # Environment config loader
│   ├── database/          # DB connection & auto-migration
│   ├── handlers/          # HTTP handlers (thin — delegate to services)
│   ├── middleware/         # JWT auth, CORS, logger
│   ├── models/            # GORM models (DB schema)
│   ├── repository/        # Data access layer (all SQL via GORM)
│   ├── services/          # Business logic + DTOs
│   └── utils/             # JWT helpers
├── pkg/
│   └── response/          # Standard JSON response envelope
├── .env.example
├── Makefile
└── go.mod
```

---

## Quick Start

```bash
# 1. Clone & install dependencies
go mod tidy

# 2. Start Postgres (Docker)
make docker-up

# 3. Copy and configure env
cp .env.example .env

# 4. Run the server
make run
```

---

## API Reference

All responses follow the shape:
```json
{ "success": true, "message": "...", "data": { ... } }
{ "success": false, "error": "..." }
```

### Auth (public)

| Method | Path                    | Body                              | Description        |
|--------|-------------------------|-----------------------------------|--------------------|
| POST   | `/api/v1/auth/signup`   | `{ name, email, password }`       | Create account     |
| POST   | `/api/v1/auth/login`    | `{ email, password }`             | Login → JWT token  |

### Users (protected — `Authorization: Bearer <token>`)

| Method | Path               | Body           | Description            |
|--------|--------------------|----------------|------------------------|
| GET    | `/api/v1/users/me` | —              | Get own profile        |
| PUT    | `/api/v1/users/me` | `{ name }`     | Update name            |
| DELETE | `/api/v1/users/me` | —              | Soft-delete account    |

---

## Design Decisions

- **Repository pattern** — all DB calls are behind interfaces, easy to mock in tests.
- **Service layer** — business logic is isolated from HTTP concerns.
- **UUIDs** as primary keys — avoids sequential ID enumeration.
- **Soft deletes** via GORM's `DeletedAt` — records are never hard-deleted.
- **bcrypt** password hashing — `Password` field is excluded from JSON serialisation.
- **Graceful shutdown** — in-flight requests finish before the process exits.
- **Connection pool** — `SetMaxOpenConns` / `SetMaxIdleConns` tuned for moderate load.
