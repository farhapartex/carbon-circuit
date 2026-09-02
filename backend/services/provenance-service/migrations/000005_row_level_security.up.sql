CREATE OR REPLACE FUNCTION provenance.current_organization_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.organization_id', true), '')::uuid
$$;

CREATE OR REPLACE FUNCTION provenance.current_user_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.user_id', true), '')::uuid
$$;

ALTER TABLE provenance.batch_parents ENABLE ROW LEVEL SECURITY;
ALTER TABLE provenance.batch_parents FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON provenance.batch_parents
    USING (organization_id = provenance.current_organization_id())
    WITH CHECK (organization_id = provenance.current_organization_id());

ALTER TABLE provenance.batches ENABLE ROW LEVEL SECURITY;
ALTER TABLE provenance.batches FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON provenance.batches
    USING (
        organization_id = provenance.current_organization_id()
        OR EXISTS (
            SELECT 1
            FROM provenance.batch_parents declaration
            WHERE declaration.parent_batch_id = batches.id
              AND declaration.organization_id = provenance.current_organization_id()
              AND declaration.deleted_at IS NULL
        )
    )
    WITH CHECK (organization_id = provenance.current_organization_id());

ALTER TABLE provenance.checkpoints ENABLE ROW LEVEL SECURITY;
ALTER TABLE provenance.checkpoints FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON provenance.checkpoints
    USING (
        organization_id = provenance.current_organization_id()
        OR EXISTS (
            SELECT 1
            FROM provenance.batch_parents declaration
            WHERE declaration.parent_batch_id = checkpoints.batch_id
              AND declaration.organization_id = provenance.current_organization_id()
              AND declaration.deleted_at IS NULL
        )
    )
    WITH CHECK (organization_id = provenance.current_organization_id());

ALTER TABLE provenance.idempotency_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE provenance.idempotency_records FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON provenance.idempotency_records
    USING (
        organization_id = provenance.current_organization_id()
        OR (organization_id IS NULL AND user_id = provenance.current_user_id())
    )
    WITH CHECK (
        organization_id = provenance.current_organization_id()
        OR (organization_id IS NULL AND user_id = provenance.current_user_id())
    );
