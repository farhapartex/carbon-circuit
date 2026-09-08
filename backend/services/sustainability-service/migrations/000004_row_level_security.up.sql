CREATE OR REPLACE FUNCTION sustainability.current_organization_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.organization_id', true), '')::uuid
$$;

CREATE OR REPLACE FUNCTION sustainability.current_user_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.user_id', true), '')::uuid
$$;

ALTER TABLE sustainability.claims ENABLE ROW LEVEL SECURITY;
ALTER TABLE sustainability.claims FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sustainability.claims
    USING (organization_id = sustainability.current_organization_id())
    WITH CHECK (organization_id = sustainability.current_organization_id());

ALTER TABLE sustainability.claim_evidence ENABLE ROW LEVEL SECURITY;
ALTER TABLE sustainability.claim_evidence FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sustainability.claim_evidence
    USING (organization_id = sustainability.current_organization_id())
    WITH CHECK (organization_id = sustainability.current_organization_id());

ALTER TABLE sustainability.idempotency_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE sustainability.idempotency_records FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sustainability.idempotency_records
    USING (
        organization_id = sustainability.current_organization_id()
        OR (organization_id IS NULL AND user_id = sustainability.current_user_id())
    )
    WITH CHECK (
        organization_id = sustainability.current_organization_id()
        OR (organization_id IS NULL AND user_id = sustainability.current_user_id())
    );
