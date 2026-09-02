CREATE TYPE provenance_read.checkpoint_type AS ENUM (
    'production_complete', 'departed_origin', 'customs_export',
    'customs_import', 'arrived_destination'
);

CREATE TYPE provenance_read.anchor_status AS ENUM (
    'unanchored', 'provisional', 'confirmed'
);

CREATE TABLE provenance_read.public_batches (
    id                         uuid PRIMARY KEY,
    created_at                 timestamptz NOT NULL DEFAULT now(),
    updated_at                 timestamptz NOT NULL DEFAULT now(),
    deleted_at                 timestamptz,
    version                    integer NOT NULL DEFAULT 1,
    public_reference           char(22) NOT NULL,
    product_category           text NOT NULL,
    component_type             text NOT NULL,
    originating_facility_name  text NOT NULL,
    originating_facility_country char(2) NOT NULL,
    produced_at                timestamptz NOT NULL,
    provenance_score           integer NOT NULL DEFAULT 0,
    score_components           json NOT NULL DEFAULT '[]'::json,
    last_updated_at            timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT public_batches_score_range CHECK (provenance_score BETWEEN 0 AND 100)
);

CREATE UNIQUE INDEX public_batches_reference
    ON provenance_read.public_batches (public_reference);

CREATE TABLE provenance_read.public_checkpoints (
    id                      uuid PRIMARY KEY,
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    deleted_at              timestamptz,
    version                 integer NOT NULL DEFAULT 1,
    batch_id                uuid NOT NULL REFERENCES provenance_read.public_batches (id),
    type                    provenance_read.checkpoint_type NOT NULL,
    location_label          text NOT NULL,
    country_code            char(2) NOT NULL,
    shipping_method         text,
    occurred_at             timestamptz NOT NULL,
    anchor_status           provenance_read.anchor_status NOT NULL DEFAULT 'unanchored',
    anchor_epoch            integer,
    anchor_transaction_hash char(66),
    superseded              boolean NOT NULL DEFAULT false
);

CREATE INDEX public_checkpoints_batch_timeline
    ON provenance_read.public_checkpoints (batch_id, occurred_at, id);

CREATE TABLE provenance_read.inbox_events (
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

CREATE UNIQUE INDEX inbox_events_event_id
    ON provenance_read.inbox_events (event_id, consumer_group);
CREATE INDEX inbox_events_sweep ON provenance_read.inbox_events (processed_at);
