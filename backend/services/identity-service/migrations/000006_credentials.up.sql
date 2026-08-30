CREATE TABLE identity.api_keys (
    id                   uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz,
    version              integer NOT NULL DEFAULT 1,
    organization_id      uuid NOT NULL REFERENCES identity.organizations (id),
    name                 text NOT NULL,
    prefix               char(8) NOT NULL,
    secret_hmac          bytea NOT NULL,
    created_by_user_id   uuid NOT NULL REFERENCES identity.users (id),
    last_used_at         timestamptz,
    revoked_at           timestamptz,
    revoked_by_user_id   uuid REFERENCES identity.users (id)
);

CREATE UNIQUE INDEX api_keys_prefix ON identity.api_keys (prefix);
CREATE INDEX api_keys_organization_id ON identity.api_keys (organization_id);
CREATE INDEX api_keys_created_by_user_id ON identity.api_keys (created_by_user_id);
CREATE INDEX api_keys_revoked_by_user_id ON identity.api_keys (revoked_by_user_id);

CREATE TABLE identity.sessions (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz,
    version       integer NOT NULL DEFAULT 1,
    user_id       uuid NOT NULL REFERENCES identity.users (id),
    user_agent    text NOT NULL,
    ip_address    inet NOT NULL,
    started_at    timestamptz NOT NULL DEFAULT now(),
    last_seen_at  timestamptz NOT NULL DEFAULT now(),
    revoked_at    timestamptz
);

CREATE INDEX sessions_user_id ON identity.sessions (user_id);
CREATE INDEX sessions_user_active ON identity.sessions (user_id, last_seen_at DESC) WHERE revoked_at IS NULL;

CREATE TABLE identity.token_revocations (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz,
    version     integer NOT NULL DEFAULT 1,
    subject     text NOT NULL,
    reason      identity.revocation_reason NOT NULL,
    revoked_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz NOT NULL
);

CREATE INDEX token_revocations_subject ON identity.token_revocations (subject);
CREATE INDEX token_revocations_expiry ON identity.token_revocations (expires_at);
