DROP POLICY IF EXISTS tenant_isolation ON evidence.idempotency_records;
ALTER TABLE evidence.idempotency_records NO FORCE ROW LEVEL SECURITY;
ALTER TABLE evidence.idempotency_records DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON evidence.documents;
ALTER TABLE evidence.documents NO FORCE ROW LEVEL SECURITY;
ALTER TABLE evidence.documents DISABLE ROW LEVEL SECURITY;
DROP FUNCTION IF EXISTS evidence.current_user_id();
DROP FUNCTION IF EXISTS evidence.current_organization_id();
