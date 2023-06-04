CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT NOW(),
    activated boolean NOT NULL,
    version integer NOT NULL DEFAULT 1
)
