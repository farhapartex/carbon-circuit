CREATE OR REPLACE FUNCTION sustainability.reviewable_by_current_verifier(subject uuid)
RETURNS boolean
LANGUAGE sql STABLE AS $$
    SELECT sustainability.current_platform_role() = 'verifier'
       AND NOT EXISTS (
            SELECT 1
            FROM sustainability.verifier_exclusions exclusion
            WHERE exclusion.verifier_user_id = sustainability.current_user_id()
              AND exclusion.organization_id = subject
              AND exclusion.deleted_at IS NULL
       )
$$;
