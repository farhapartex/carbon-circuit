CREATE TYPE sustainability.activity_type AS ENUM (
    'renewable_energy', 'reduced_emission_logistics', 'responsible_sourcing'
);

CREATE TYPE sustainability.claim_status AS ENUM (
    'submitted', 'ai_review', 'human_review', 'approved', 'rejected', 'more_information_requested'
);

CREATE TYPE sustainability.queue_priority AS ENUM ('normal', 'high', 'critical');

CREATE TYPE sustainability.idempotency_state AS ENUM ('processing', 'completed', 'failed');

CREATE TABLE sustainability.claims (
    id                        uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now(),
    deleted_at                timestamptz,
    version                   integer NOT NULL DEFAULT 1,
    organization_id           uuid NOT NULL,
    submitted_by_user_id      uuid NOT NULL,
    facility_id               uuid NOT NULL,
    facility_name             text NOT NULL,
    activity_type             sustainability.activity_type NOT NULL,
    vintage_year              integer NOT NULL,
    period_start              date NOT NULL,
    period_end                date NOT NULL,
    declared_figures          jsonb NOT NULL,
    requested_amount          numeric(28,6) NOT NULL,
    computed_ceiling          numeric(28,6) NOT NULL,
    capacity_basis            numeric(28,6) NOT NULL,
    capacity_source           text NOT NULL,
    discount_factor           numeric(3,2) NOT NULL,
    reference_factor_id       uuid NOT NULL REFERENCES sustainability.reference_factors (id),
    reference_factor_value    numeric(20,6) NOT NULL,
    status                    sustainability.claim_status NOT NULL DEFAULT 'submitted',
    priority                  sustainability.queue_priority NOT NULL DEFAULT 'normal',
    requires_dual_approval    boolean NOT NULL DEFAULT false,
    exclusivity_attested_at   timestamptz NOT NULL,
    exclusivity_attested_by   uuid NOT NULL,
    issued_amount             numeric(28,6),
    CONSTRAINT claims_period_ordered CHECK (period_end >= period_start),
    CONSTRAINT claims_requested_positive CHECK (requested_amount > 0),
    CONSTRAINT claims_ceiling_positive CHECK (computed_ceiling > 0),
    CONSTRAINT claims_capacity_positive CHECK (capacity_basis > 0),
    CONSTRAINT claims_discount_known CHECK (discount_factor IN (1.00, 0.75, 0.50)),
    CONSTRAINT claims_vintage_sane CHECK (vintage_year BETWEEN 2020 AND 2100),
    CONSTRAINT claims_issued_within_ceiling CHECK (issued_amount IS NULL OR issued_amount <= computed_ceiling)
);

CREATE INDEX claims_organization_listing
    ON sustainability.claims (organization_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX claims_queue
    ON sustainability.claims (priority DESC, created_at)
    WHERE deleted_at IS NULL AND status IN ('submitted', 'ai_review', 'human_review');

CREATE INDEX claims_facility_vintage
    ON sustainability.claims (facility_id, vintage_year, activity_type)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN sustainability.claims.computed_ceiling IS
    'The hard cap from PRD 2.1, computed from capacity_basis and never from requested_amount. Nothing may issue above it.';
COMMENT ON COLUMN sustainability.claims.reference_factor_id IS
    'The exact reference row used, pinned forever so a recomputation years later produces the same answer.';

CREATE TABLE sustainability.claim_evidence (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    version         integer NOT NULL DEFAULT 1,
    organization_id uuid NOT NULL,
    claim_id        uuid NOT NULL REFERENCES sustainability.claims (id),
    evidence_id     uuid NOT NULL,
    content_hash    char(64) NOT NULL,
    file_name       text NOT NULL,
    media_type      text NOT NULL,
    page_count      integer
);

CREATE UNIQUE INDEX claim_evidence_unique
    ON sustainability.claim_evidence (claim_id, evidence_id)
    WHERE deleted_at IS NULL;

CREATE INDEX claim_evidence_hash ON sustainability.claim_evidence (content_hash);

COMMENT ON COLUMN sustainability.claim_evidence.content_hash IS
    'Copied from Evidence Service at submission so the duplicate-evidence rule in PRD 3.6 can compare across claims without reaching into another schema.';

CREATE TABLE sustainability.idempotency_records (
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
    state            sustainability.idempotency_state NOT NULL DEFAULT 'processing',
    response_status  integer,
    response_body    json,
    resource_id      uuid,
    completed_at     timestamptz,
    CONSTRAINT idempotency_records_scope
        CHECK (organization_id IS NOT NULL OR user_id IS NOT NULL)
);

CREATE UNIQUE INDEX idempotency_records_tenant_key
    ON sustainability.idempotency_records (organization_id, endpoint, idempotency_key)
    WHERE organization_id IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX idempotency_records_user_key
    ON sustainability.idempotency_records (user_id, endpoint, idempotency_key)
    WHERE organization_id IS NULL AND deleted_at IS NULL;
CREATE INDEX idempotency_records_sweep ON sustainability.idempotency_records (created_at);

CREATE TABLE sustainability.outbox_events (
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

CREATE INDEX outbox_events_aggregate_id ON sustainability.outbox_events (aggregate_id);
CREATE INDEX outbox_events_unpublished
    ON sustainability.outbox_events (created_at)
    WHERE published_at IS NULL;

CREATE TABLE sustainability.inbox_events (
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

CREATE UNIQUE INDEX inbox_events_event_id ON sustainability.inbox_events (event_id);
CREATE INDEX inbox_events_sweep ON sustainability.inbox_events (processed_at);
