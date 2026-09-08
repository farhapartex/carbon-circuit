CREATE OR REPLACE FUNCTION evidence.current_organization_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.organization_id', true), '')::uuid
$$;

CREATE OR REPLACE FUNCTION evidence.current_user_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.user_id', true), '')::uuid
$$;

ALTER TABLE evidence.documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE evidence.documents FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON evidence.documents
    USING (organization_id = evidence.current_organization_id())
    WITH CHECK (organization_id = evidence.current_organization_id());

ALTER TABLE evidence.idempotency_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE evidence.idempotency_records FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON evidence.idempotency_records
    USING (
        organization_id = evidence.current_organization_id()
        OR (organization_id IS NULL AND user_id = evidence.current_user_id())
    )
    WITH CHECK (
        organization_id = evidence.current_organization_id()
        OR (organization_id IS NULL AND user_id = evidence.current_user_id())
    );
