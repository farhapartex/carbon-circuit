CREATE TYPE evidence.scan_status AS ENUM ('pending', 'clean', 'failed');

CREATE TYPE evidence.scan_verdict AS ENUM (
    'accepted',
    'rejected_media_type',
    'rejected_content_mismatch',
    'rejected_too_large',
    'rejected_page_count',
    'rejected_active_content',
    'rejected_malware',
    'rejected_unreadable'
);

CREATE TYPE evidence.document_purpose AS ENUM ('claim_evidence', 'batch_certification');

CREATE TYPE evidence.idempotency_state AS ENUM ('processing', 'completed', 'failed');

CREATE TABLE evidence.documents (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz,
    version             integer NOT NULL DEFAULT 1,
    organization_id     uuid NOT NULL,
    uploaded_by_user_id uuid NOT NULL,
    purpose             evidence.document_purpose NOT NULL,
    file_name           text NOT NULL,
    declared_media_type text NOT NULL,
    detected_media_type text NOT NULL,
    byte_size           bigint NOT NULL,
    page_count          integer,
    content_hash        char(64) NOT NULL,
    storage_key         text,
    scan_status         evidence.scan_status NOT NULL DEFAULT 'pending',
    scan_verdict        evidence.scan_verdict NOT NULL,
    scanned_by          text NOT NULL,
    scanned_at          timestamptz NOT NULL DEFAULT now(),
    active_content_stripped boolean NOT NULL DEFAULT false,
    CONSTRAINT documents_stored_only_when_clean CHECK (
        (scan_status = 'clean') = (storage_key IS NOT NULL)
    ),
    CONSTRAINT documents_verdict_matches_status CHECK (
        (scan_status = 'clean') = (scan_verdict = 'accepted')
    ),
    CONSTRAINT documents_byte_size_positive CHECK (byte_size > 0),
    CONSTRAINT documents_page_count_positive CHECK (page_count IS NULL OR page_count > 0)
);

CREATE UNIQUE INDEX documents_storage_key ON evidence.documents (storage_key);
CREATE INDEX documents_organization_listing
    ON evidence.documents (organization_id, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX documents_content_hash
    ON evidence.documents (content_hash)
    WHERE scan_status = 'clean' AND deleted_at IS NULL;

CREATE TABLE evidence.idempotency_records (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    organization_id  uuid,
    user_id          uuid,
    endpoint         text NOT NULL,
    idempotency_key  text NOT NULL,
    request_hash     bytea NOT NULL,
    state            evidence.idempotency_state NOT NULL DEFAULT 'processing',
    response_status  integer,
    response_body    json,
    resource_id      uuid,
    completed_at     timestamptz,
    CONSTRAINT idempotency_records_scope
        CHECK (organization_id IS NOT NULL OR user_id IS NOT NULL)
);

CREATE UNIQUE INDEX idempotency_records_tenant_key
    ON evidence.idempotency_records (organization_id, endpoint, idempotency_key)
    WHERE organization_id IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX idempotency_records_user_key
    ON evidence.idempotency_records (user_id, endpoint, idempotency_key)
    WHERE organization_id IS NULL AND deleted_at IS NULL;

CREATE INDEX idempotency_records_user_id ON evidence.idempotency_records (user_id);
CREATE INDEX idempotency_records_sweep ON evidence.idempotency_records (created_at);

CREATE TABLE evidence.outbox_events (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    version         integer NOT NULL DEFAULT 1,
    aggregate_type  text NOT NULL,
    aggregate_id    uuid NOT NULL,
    event_type      text NOT NULL,
    payload         jsonb NOT NULL,
    headers         jsonb NOT NULL DEFAULT '{}'::jsonb,
    published_at    timestamptz
);

CREATE INDEX outbox_events_aggregate_id ON evidence.outbox_events (aggregate_id);
CREATE INDEX outbox_events_unpublished
    ON evidence.outbox_events (created_at)
    WHERE published_at IS NULL;
