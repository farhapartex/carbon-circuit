CREATE TABLE provenance.batches (
    id                         uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at                 timestamptz NOT NULL DEFAULT now(),
    updated_at                 timestamptz NOT NULL DEFAULT now(),
    deleted_at                 timestamptz,
    version                    integer NOT NULL DEFAULT 1,
    organization_id            uuid NOT NULL,
    originating_facility_id    uuid NOT NULL,
    originating_facility_name  text NOT NULL,
    public_reference           char(22) NOT NULL,
    product_category           provenance.product_category NOT NULL,
    component_type             text NOT NULL,
    lot_number                 text,
    quantity                   numeric(28,6) NOT NULL,
    unit                       text NOT NULL,
    produced_at                timestamptz NOT NULL,
    external_id                text,
    checkpoint_count           integer NOT NULL DEFAULT 0,
    provenance_score           integer NOT NULL DEFAULT 0,
    score_components           json NOT NULL DEFAULT '[]'::json,
    CONSTRAINT batches_quantity_positive CHECK (quantity > 0),
    CONSTRAINT batches_score_range CHECK (provenance_score BETWEEN 0 AND 100),
    CONSTRAINT batches_reference_base62 CHECK (public_reference ~ '^[0-9A-Za-z]{22}$')
);

CREATE UNIQUE INDEX batches_public_reference
    ON provenance.batches (public_reference);

CREATE UNIQUE INDEX batches_external_id
    ON provenance.batches (organization_id, external_id)
    WHERE external_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX batches_organization_id ON provenance.batches (organization_id);

CREATE INDEX batches_organization_listing
    ON provenance.batches (organization_id, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX batches_originating_facility
    ON provenance.batches (originating_facility_id)
    WHERE deleted_at IS NULL;

CREATE TABLE provenance.batch_parents (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz,
    version             integer NOT NULL DEFAULT 1,
    organization_id     uuid NOT NULL,
    batch_id            uuid NOT NULL REFERENCES provenance.batches (id),
    declared_reference  char(22) NOT NULL,
    parent_batch_id     uuid REFERENCES provenance.batches (id),
    CONSTRAINT batch_parents_not_self CHECK (parent_batch_id IS NULL OR parent_batch_id <> batch_id)
);

CREATE UNIQUE INDEX batch_parents_unique_declaration
    ON provenance.batch_parents (batch_id, declared_reference)
    WHERE deleted_at IS NULL;

CREATE INDEX batch_parents_batch_id ON provenance.batch_parents (batch_id);
CREATE INDEX batch_parents_parent_batch_id ON provenance.batch_parents (parent_batch_id);
CREATE INDEX batch_parents_organization_id ON provenance.batch_parents (organization_id);
