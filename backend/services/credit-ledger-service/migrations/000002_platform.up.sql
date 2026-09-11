CREATE TABLE credit_ledger.inbox_events (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    version         integer NOT NULL DEFAULT 1,
    event_id        uuid NOT NULL,
    topic           text NOT NULL,
    consumer_group  text NOT NULL,
    processed_at    timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX inbox_events_event_id ON credit_ledger.inbox_events (event_id);
CREATE INDEX inbox_events_sweep ON credit_ledger.inbox_events (processed_at);

CREATE TABLE credit_ledger.outbox_events (
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

CREATE INDEX outbox_events_aggregate_id ON credit_ledger.outbox_events (aggregate_id);
CREATE INDEX outbox_events_unpublished
    ON credit_ledger.outbox_events (created_at)
    WHERE published_at IS NULL;

CREATE TABLE credit_ledger.idempotency_records (
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
    state            credit_ledger.idempotency_state NOT NULL DEFAULT 'processing',
    response_status  integer,
    response_body    json,
    resource_id      uuid,
    completed_at     timestamptz,
    CONSTRAINT idempotency_records_scope
        CHECK (organization_id IS NOT NULL OR user_id IS NOT NULL)
);

CREATE UNIQUE INDEX idempotency_records_tenant_key
    ON credit_ledger.idempotency_records (organization_id, endpoint, idempotency_key)
    WHERE organization_id IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX idempotency_records_user_key
    ON credit_ledger.idempotency_records (user_id, endpoint, idempotency_key)
    WHERE organization_id IS NULL AND deleted_at IS NULL;
CREATE INDEX idempotency_records_sweep ON credit_ledger.idempotency_records (created_at);
