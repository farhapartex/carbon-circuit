ALTER TABLE evidence.documents
    DROP CONSTRAINT documents_byte_size_recorded;

ALTER TABLE evidence.documents
    ADD CONSTRAINT documents_byte_size_positive CHECK (byte_size > 0);
