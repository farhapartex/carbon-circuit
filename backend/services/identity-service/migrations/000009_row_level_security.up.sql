CREATE OR REPLACE FUNCTION identity.current_organization_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.organization_id', true), '')::uuid
$$;

CREATE OR REPLACE FUNCTION identity.current_user_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.user_id', true), '')::uuid
$$;

ALTER TABLE identity.organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.organizations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_read ON identity.organizations FOR SELECT
    USING (id = identity.current_organization_id());
CREATE POLICY tenant_write ON identity.organizations FOR UPDATE
    USING (id = identity.current_organization_id());
CREATE POLICY registration_insert ON identity.organizations FOR INSERT
    WITH CHECK (true);

ALTER TABLE identity.organization_memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.organization_memberships FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.organization_memberships
    USING (
        organization_id = identity.current_organization_id()
        OR user_id = identity.current_user_id()
    )
    WITH CHECK (
        organization_id = identity.current_organization_id()
        OR user_id = identity.current_user_id()
    );

ALTER TABLE identity.facilities ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.facilities FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.facilities
    USING (organization_id = identity.current_organization_id())
    WITH CHECK (organization_id = identity.current_organization_id());

ALTER TABLE identity.invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.invitations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.invitations
    USING (organization_id = identity.current_organization_id())
    WITH CHECK (organization_id = identity.current_organization_id());

ALTER TABLE identity.treasury_addresses ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.treasury_addresses FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.treasury_addresses
    USING (organization_id = identity.current_organization_id())
    WITH CHECK (organization_id = identity.current_organization_id());

ALTER TABLE identity.treasury_address_changes ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.treasury_address_changes FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.treasury_address_changes
    USING (organization_id = identity.current_organization_id())
    WITH CHECK (organization_id = identity.current_organization_id());

ALTER TABLE identity.api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.api_keys FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.api_keys
    USING (organization_id = identity.current_organization_id())
    WITH CHECK (organization_id = identity.current_organization_id());

ALTER TABLE identity.data_export_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.data_export_requests FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.data_export_requests
    USING (organization_id = identity.current_organization_id())
    WITH CHECK (organization_id = identity.current_organization_id());

ALTER TABLE identity.deletion_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.deletion_requests FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.deletion_requests
    USING (organization_id = identity.current_organization_id())
    WITH CHECK (organization_id = identity.current_organization_id());

ALTER TABLE identity.idempotency_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.idempotency_records FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON identity.idempotency_records
    USING (organization_id = identity.current_organization_id())
    WITH CHECK (organization_id = identity.current_organization_id());
