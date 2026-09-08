ALTER TABLE evidence.documents
    ADD COLUMN stored_hash char(64);

ALTER TABLE evidence.documents
    ADD CONSTRAINT documents_stored_hash_when_clean CHECK (
        (scan_status = 'clean') = (stored_hash IS NOT NULL)
    );

COMMENT ON COLUMN evidence.documents.content_hash IS
    'SHA-256 of the bytes as submitted. This is the hash recorded publicly and the one the duplicate-evidence rule compares, so it must describe what the uploader provided.';

COMMENT ON COLUMN evidence.documents.stored_hash IS
    'SHA-256 of the sanitized bytes actually written to object storage. Sanitization is not byte-deterministic, so this can never stand in for content_hash.';
