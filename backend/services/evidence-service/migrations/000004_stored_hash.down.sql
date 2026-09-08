ALTER TABLE evidence.documents
    DROP CONSTRAINT documents_stored_hash_when_clean;

ALTER TABLE evidence.documents
    DROP COLUMN stored_hash;
