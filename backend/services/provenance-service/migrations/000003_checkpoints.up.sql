CREATE TABLE provenance.checkpoints (
    id                             uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at                     timestamptz NOT NULL DEFAULT now(),
    updated_at                     timestamptz NOT NULL DEFAULT now(),
    deleted_at                     timestamptz,
    version                        integer NOT NULL DEFAULT 1,
    organization_id                uuid NOT NULL,
    batch_id                       uuid NOT NULL REFERENCES provenance.batches (id),
    type                           provenance.checkpoint_type NOT NULL,
    location_label                 text NOT NULL,
    country_code                   char(2) NOT NULL,
    latitude                       numeric(9,6),
    longitude                      numeric(9,6),
    shipping_method                provenance.shipping_method,
    occurred_at                    timestamptz NOT NULL,
    reported_at                    timestamptz NOT NULL DEFAULT now(),
    reported_by_organization_id    uuid NOT NULL,
    reported_by_organization_name  text NOT NULL,
    anchor_status                  provenance.anchor_status NOT NULL DEFAULT 'unanchored',
    anchor_epoch                   integer,
    anchor_transaction_hash        char(66),
    inclusion_proof_available      boolean NOT NULL DEFAULT false,
    supersedes_checkpoint_id       uuid REFERENCES provenance.checkpoints (id),
    superseded_by_checkpoint_id    uuid REFERENCES provenance.checkpoints (id),
    correction_reason              text,
    external_id                    text,
    CONSTRAINT checkpoints_movement_has_method CHECK (
        type = 'production_complete' OR shipping_method IS NOT NULL
    ),
    CONSTRAINT checkpoints_coordinates_paired CHECK (
        (latitude IS NULL) = (longitude IS NULL)
    ),
    CONSTRAINT checkpoints_correction_has_reason CHECK (
        supersedes_checkpoint_id IS NULL OR correction_reason IS NOT NULL
    ),
    CONSTRAINT checkpoints_not_self_superseding CHECK (
        supersedes_checkpoint_id IS NULL OR supersedes_checkpoint_id <> id
    )
);

CREATE UNIQUE INDEX checkpoints_external_id
    ON provenance.checkpoints (organization_id, external_id)
    WHERE external_id IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX checkpoints_organization_id ON provenance.checkpoints (organization_id);
CREATE INDEX checkpoints_batch_id ON provenance.checkpoints (batch_id);
CREATE INDEX checkpoints_reported_by ON provenance.checkpoints (reported_by_organization_id);

CREATE INDEX checkpoints_batch_timeline
    ON provenance.checkpoints (batch_id, occurred_at, id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX checkpoints_one_correction_per_original
    ON provenance.checkpoints (supersedes_checkpoint_id)
    WHERE supersedes_checkpoint_id IS NOT NULL AND deleted_at IS NULL;
