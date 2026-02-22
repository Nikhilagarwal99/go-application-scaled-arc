DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";