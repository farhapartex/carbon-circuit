DROP INDEX IF EXISTS provenance_read.public_batches_projected;
ALTER TABLE provenance_read.public_batches DROP COLUMN IF EXISTS projected;
