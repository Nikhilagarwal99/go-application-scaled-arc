CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id              UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(100)    NOT NULL,
    email           VARCHAR(150)    NOT NULL,
    password        VARCHAR(255)    NOT NULL,
    email_verified  BOOLEAN         NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ     -- nullable, used for soft deletes
);

-- Unique index on email — only among non-deleted users
-- A deleted user's email should be reusable
CREATE UNIQUE INDEX idx_users_email
    ON users (email)
    WHERE deleted_at IS NULL;

-- Index on deleted_at — GORM always adds WHERE deleted_at IS NULL
-- to every query, so this makes those queries fast
CREATE INDEX idx_users_deleted_at
    ON users (deleted_at);