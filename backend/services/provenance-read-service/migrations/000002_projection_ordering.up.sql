ALTER TABLE provenance_read.public_batches
    ADD COLUMN projected boolean NOT NULL DEFAULT false;

UPDATE provenance_read.public_batches SET projected = true;

ALTER TABLE provenance_read.public_batches
    ALTER COLUMN product_category DROP NOT NULL,
    ALTER COLUMN component_type DROP NOT NULL,
    ALTER COLUMN originating_facility_name DROP NOT NULL,
    ALTER COLUMN originating_facility_country DROP NOT NULL,
    ALTER COLUMN produced_at DROP NOT NULL;

CREATE INDEX public_batches_projected
    ON provenance_read.public_batches (public_reference)
    WHERE projected;
