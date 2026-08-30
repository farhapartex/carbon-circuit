DROP POLICY IF EXISTS tenant_isolation ON identity.idempotency_records;
DROP POLICY IF EXISTS tenant_isolation ON identity.deletion_requests;
DROP POLICY IF EXISTS tenant_isolation ON identity.data_export_requests;
DROP POLICY IF EXISTS tenant_isolation ON identity.api_keys;
DROP POLICY IF EXISTS tenant_isolation ON identity.treasury_address_changes;
DROP POLICY IF EXISTS tenant_isolation ON identity.treasury_addresses;
DROP POLICY IF EXISTS tenant_isolation ON identity.invitations;
DROP POLICY IF EXISTS tenant_isolation ON identity.facilities;
DROP POLICY IF EXISTS tenant_isolation ON identity.organization_memberships;
DROP POLICY IF EXISTS registration_insert ON identity.organizations;
DROP POLICY IF EXISTS tenant_write ON identity.organizations;
DROP POLICY IF EXISTS tenant_read ON identity.organizations;

ALTER TABLE identity.idempotency_records DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.deletion_requests DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.data_export_requests DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.api_keys DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.treasury_address_changes DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.treasury_addresses DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.invitations DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.facilities DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.organization_memberships DISABLE ROW LEVEL SECURITY;
ALTER TABLE identity.organizations DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS identity.current_user_id();
DROP FUNCTION IF EXISTS identity.current_organization_id();
