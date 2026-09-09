CREATE OR REPLACE FUNCTION sustainability.reviewable_by_current_verifier(subject uuid)
RETURNS boolean
LANGUAGE sql STABLE AS $$
    SELECT sustainability.current_platform_role() = 'verifier'
       AND sustainability.current_user_id() IS NOT NULL
       AND NOT EXISTS (
            SELECT 1
            FROM sustainability.verifier_exclusions exclusion
            WHERE exclusion.verifier_user_id = sustainability.current_user_id()
              AND exclusion.organization_id = subject
              AND exclusion.deleted_at IS NULL
       )
$$;

COMMENT ON FUNCTION sustainability.reviewable_by_current_verifier(uuid) IS
    'True only when the caller holds the verifier platform role for this transaction, is identified, and has not been declared related to the owning organization. The identity check is load-bearing: without it an unset user id makes every exclusion comparison null, the NOT EXISTS holds vacuously, and a verifier reads organizations they were barred from.';
