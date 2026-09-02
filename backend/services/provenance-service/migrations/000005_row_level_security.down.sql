DROP POLICY IF EXISTS tenant_isolation ON provenance.idempotency_records;
ALTER TABLE provenance.idempotency_records NO FORCE ROW LEVEL SECURITY;
ALTER TABLE provenance.idempotency_records DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON provenance.checkpoints;
ALTER TABLE provenance.checkpoints NO FORCE ROW LEVEL SECURITY;
ALTER TABLE provenance.checkpoints DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON provenance.batches;
ALTER TABLE provenance.batches NO FORCE ROW LEVEL SECURITY;
ALTER TABLE provenance.batches DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON provenance.batch_parents;
ALTER TABLE provenance.batch_parents NO FORCE ROW LEVEL SECURITY;
ALTER TABLE provenance.batch_parents DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS provenance.current_user_id();
DROP FUNCTION IF EXISTS provenance.current_organization_id();
