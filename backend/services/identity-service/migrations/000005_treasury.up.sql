CREATE TABLE identity.siwe_nonces (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz,
    version      integer NOT NULL DEFAULT 1,
    nonce        text NOT NULL,
    domain       text NOT NULL,
    user_id      uuid REFERENCES identity.users (id),
    issued_at    timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL,
    consumed_at  timestamptz
);

CREATE UNIQUE INDEX siwe_nonces_nonce ON identity.siwe_nonces (nonce);
CREATE INDEX siwe_nonces_user_id ON identity.siwe_nonces (user_id);
CREATE INDEX siwe_nonces_expiry ON identity.siwe_nonces (expires_at) WHERE consumed_at IS NULL;

CREATE TABLE identity.treasury_addresses (
    id                     uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz,
    version                integer NOT NULL DEFAULT 1,
    organization_id        uuid NOT NULL REFERENCES identity.organizations (id),
    address                char(42) NOT NULL,
    state                  identity.treasury_state NOT NULL DEFAULT 'active',
    proof_signature        bytea NOT NULL,
    nonce_id               uuid NOT NULL REFERENCES identity.siwe_nonces (id),
    designated_by_user_id  uuid NOT NULL REFERENCES identity.users (id),
    designated_at          timestamptz NOT NULL DEFAULT now(),
    superseded_at          timestamptz,
    CONSTRAINT treasury_addresses_address_shape CHECK (address ~ '^0x[0-9a-fA-F]{40}$')
);

CREATE UNIQUE INDEX treasury_addresses_one_active_per_organization
    ON identity.treasury_addresses (organization_id)
    WHERE state = 'active' AND deleted_at IS NULL;
CREATE UNIQUE INDEX treasury_addresses_nonce_id ON identity.treasury_addresses (nonce_id);
CREATE INDEX treasury_addresses_organization_id ON identity.treasury_addresses (organization_id);
CREATE INDEX treasury_addresses_designated_by_user_id ON identity.treasury_addresses (designated_by_user_id);

CREATE TABLE identity.treasury_address_changes (
    id                    uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz,
    version               integer NOT NULL DEFAULT 1,
    organization_id       uuid NOT NULL REFERENCES identity.organizations (id),
    current_address       char(42),
    requested_address     char(42) NOT NULL,
    state                 identity.treasury_change_state NOT NULL DEFAULT 'pending',
    proof_signature       bytea NOT NULL,
    mfa_verified_at       timestamptz NOT NULL,
    initiated_by_user_id  uuid NOT NULL REFERENCES identity.users (id),
    initiated_at          timestamptz NOT NULL DEFAULT now(),
    effective_at          timestamptz NOT NULL,
    resolved_at           timestamptz,
    resolved_by_user_id   uuid REFERENCES identity.users (id),
    CONSTRAINT treasury_changes_delay_is_at_least_72_hours
        CHECK (effective_at >= initiated_at + interval '72 hours')
);

CREATE UNIQUE INDEX treasury_address_changes_one_pending_per_organization
    ON identity.treasury_address_changes (organization_id)
    WHERE state = 'pending' AND deleted_at IS NULL;
CREATE INDEX treasury_address_changes_organization_id ON identity.treasury_address_changes (organization_id);
CREATE INDEX treasury_address_changes_initiated_by_user_id ON identity.treasury_address_changes (initiated_by_user_id);
CREATE INDEX treasury_address_changes_resolved_by_user_id ON identity.treasury_address_changes (resolved_by_user_id);
CREATE INDEX treasury_address_changes_due ON identity.treasury_address_changes (effective_at) WHERE state = 'pending';
