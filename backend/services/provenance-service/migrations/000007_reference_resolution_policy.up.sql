DROP FUNCTION IF EXISTS provenance.resolve_public_reference(char(22));

CREATE OR REPLACE FUNCTION provenance.resolving_reference() RETURNS text
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.resolving_reference', true), '')
$$;

CREATE POLICY reference_resolution ON provenance.batches
    FOR SELECT
    USING (public_reference = provenance.resolving_reference());
