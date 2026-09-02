CREATE OR REPLACE FUNCTION provenance.resolve_public_reference(candidate char(22))
RETURNS uuid
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = provenance, pg_temp
AS $$
    SELECT id
    FROM provenance.batches
    WHERE public_reference = candidate
      AND deleted_at IS NULL
$$;

REVOKE ALL ON FUNCTION provenance.resolve_public_reference(char(22)) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION provenance.resolve_public_reference(char(22)) TO provenance_service;
