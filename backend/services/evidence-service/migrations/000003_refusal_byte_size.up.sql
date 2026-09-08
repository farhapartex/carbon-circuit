ALTER TABLE evidence.documents
    DROP CONSTRAINT documents_byte_size_positive;

ALTER TABLE evidence.documents
    ADD CONSTRAINT documents_byte_size_recorded CHECK (byte_size >= 0);
