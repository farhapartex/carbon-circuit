CREATE TABLE identity.data_export_requests (
    id                    uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz,
    version               integer NOT NULL DEFAULT 1,
    organization_id       uuid NOT NULL REFERENCES identity.organizations (id),
    requested_by_user_id  uuid NOT NULL REFERENCES identity.users (id),
    state                 identity.export_state NOT NULL DEFAULT 'processing',
    requested_at          timestamptz NOT NULL DEFAULT now(),
    completed_at          timestamptz,
    download_token_hash   bytea,
    expires_at            timestamptz
);

CREATE INDEX data_export_requests_organization_id ON identity.data_export_requests (organization_id);
CREATE INDEX data_export_requests_requested_by_user_id ON identity.data_export_requests (requested_by_user_id);
CREATE INDEX data_export_requests_rate_limit
    ON identity.data_export_requests (organization_id, requested_at DESC);

CREATE TABLE identity.deletion_requests (
    id                    uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz,
    version               integer NOT NULL DEFAULT 1,
    organization_id       uuid NOT NULL REFERENCES identity.organizations (id),
    requested_by_user_id  uuid NOT NULL REFERENCES identity.users (id),
    state                 identity.deletion_state NOT NULL DEFAULT 'requested',
    blocked_reason        text,
    requested_at          timestamptz NOT NULL DEFAULT now(),
    purge_after           timestamptz NOT NULL,
    completed_at          timestamptz,
    CONSTRAINT deletion_requests_blocked_needs_reason
        CHECK (state <> 'blocked' OR blocked_reason IS NOT NULL)
);

CREATE UNIQUE INDEX deletion_requests_one_open_per_organization
    ON identity.deletion_requests (organization_id)
    WHERE state <> 'completed' AND deleted_at IS NULL;
CREATE INDEX deletion_requests_organization_id ON identity.deletion_requests (organization_id);
CREATE INDEX deletion_requests_requested_by_user_id ON identity.deletion_requests (requested_by_user_id);
CREATE INDEX deletion_requests_due ON identity.deletion_requests (purge_after) WHERE state = 'purging';
