CREATE TABLE identity.idempotency_records (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    organization_id  uuid NOT NULL,
    endpoint         text NOT NULL,
    idempotency_key  text NOT NULL,
    request_hash     bytea NOT NULL,
    state            identity.idempotency_state NOT NULL DEFAULT 'processing',
    response_status  integer,
    response_body    jsonb,
    resource_id      uuid,
    completed_at     timestamptz
);

CREATE UNIQUE INDEX idempotency_records_key
    ON identity.idempotency_records (organization_id, endpoint, idempotency_key);
CREATE INDEX idempotency_records_sweep ON identity.idempotency_records (created_at);

CREATE TABLE identity.inbox_events (
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

CREATE UNIQUE INDEX inbox_events_event_id ON identity.inbox_events (event_id);
CREATE INDEX inbox_events_sweep ON identity.inbox_events (processed_at);

CREATE TABLE identity.outbox_events (
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

CREATE INDEX outbox_events_aggregate_id ON identity.outbox_events (aggregate_id);
CREATE INDEX outbox_events_unpublished
    ON identity.outbox_events (created_at)
    WHERE published_at IS NULL;
