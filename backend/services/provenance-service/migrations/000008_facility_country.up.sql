ALTER TABLE provenance.batches
    ADD COLUMN originating_facility_country char(2) NOT NULL DEFAULT 'XX';

ALTER TABLE provenance.batches
    ALTER COLUMN originating_facility_country DROP DEFAULT;
