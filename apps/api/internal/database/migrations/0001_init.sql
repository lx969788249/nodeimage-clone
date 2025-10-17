-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE user_role AS ENUM ('user', 'admin', 'superadmin');
CREATE TYPE user_status AS ENUM ('active', 'suspended', 'pending');
CREATE TYPE image_status AS ENUM ('processing', 'ready', 'blocked', 'deleted');
CREATE TYPE report_status AS ENUM ('pending', 'reviewing', 'resolved', 'rejected');

CREATE TABLE users (
    id                CHAR(27) PRIMARY KEY,
    email             CITEXT UNIQUE NOT NULL,
    password_hash     BYTEA NOT NULL,
    display_name      TEXT NOT NULL,
    role              user_role NOT NULL DEFAULT 'user',
    status            user_status NOT NULL DEFAULT 'active',
    avatar_url        TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

CREATE TABLE user_sessions (
    id                  CHAR(27) PRIMARY KEY,
    user_id             CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id           VARCHAR(64) NOT NULL,
    device_name         TEXT NOT NULL,
    refresh_token_hash  BYTEA NOT NULL,
    ip_address          INET,
    user_agent          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, device_id)
);

CREATE INDEX idx_user_sessions_user ON user_sessions (user_id);
CREATE INDEX idx_user_sessions_expires ON user_sessions (expires_at);

CREATE TABLE api_keys (
    id              CHAR(27) PRIMARY KEY,
    user_id         CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    key_hash        BYTEA NOT NULL,
    scopes          TEXT[] NOT NULL,
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE images (
    id             CHAR(27) PRIMARY KEY,
    user_id        CHAR(27) REFERENCES users(id) ON DELETE SET NULL,
    bucket         TEXT NOT NULL,
    object_key     TEXT NOT NULL,
    format         TEXT NOT NULL,
    width          INT NOT NULL,
    height         INT NOT NULL,
    frames         INT NOT NULL DEFAULT 1,
    size_bytes     BIGINT NOT NULL,
    nsfw_score     REAL,
    visibility     TEXT NOT NULL DEFAULT 'public',
    status         image_status NOT NULL DEFAULT 'processing',
    checksum       BYTEA NOT NULL,
    signature      BYTEA NOT NULL,
    expire_at      TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_images_user ON images (user_id);
CREATE INDEX idx_images_status ON images (status);
CREATE INDEX idx_images_expire ON images (expire_at);

CREATE TABLE image_variants (
    id           CHAR(27) PRIMARY KEY,
    image_id     CHAR(27) NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    variant      TEXT NOT NULL,
    bucket       TEXT NOT NULL,
    object_key   TEXT NOT NULL,
    format       TEXT NOT NULL,
    width        INT,
    height       INT,
    size_bytes   BIGINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (image_id, variant)
);

CREATE TABLE image_audit_logs (
    id            BIGSERIAL PRIMARY KEY,
    image_id      CHAR(27) NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    reviewer_id   CHAR(27) REFERENCES users(id),
    action        TEXT NOT NULL,
    reason        TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE reports (
    id            BIGSERIAL PRIMARY KEY,
    image_id      CHAR(27) NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    reporter_id   CHAR(27) REFERENCES users(id),
    reason        TEXT NOT NULL,
    notes         TEXT,
    status        report_status NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    handled_at    TIMESTAMPTZ
);

CREATE TABLE webhooks (
    id          CHAR(27) PRIMARY KEY,
    user_id     CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    secret      BYTEA NOT NULL,
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE failed_jobs (
    id            BIGSERIAL PRIMARY KEY,
    queue         TEXT NOT NULL,
    payload       JSONB NOT NULL,
    error_message TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS failed_jobs;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS image_audit_logs;
DROP TABLE IF EXISTS image_variants;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS image_status;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
