CREATE TABLE IF NOT EXISTS users (
    "id" serial PRIMARY KEY,
    "name" text NOT NULL,
    "email" CITEXT UNIQUE NOT NULL,
    "phone_number" text,
    "id_code" bigint UNIQUE,
    "birth_date" date,
    "password" bytea NOT NULL,
    "role" text NOT NULL,
    "class_id" integer,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "active" boolean NOT NULL DEFAULT TRUE,
    "archived" boolean NOT NULL DEFAULT FALSE
);

CREATE INDEX trgm_idx_users_name ON users USING gin (name gin_trgm_ops);

ALTER TABLE users
    ADD CONSTRAINT student_class_id_not_null CHECK ( CASE WHEN ROLE = 'student' THEN
        class_id IS NOT NULL
    END);

