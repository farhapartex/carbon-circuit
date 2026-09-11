CREATE OR REPLACE FUNCTION identity.resolving_organization() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.resolving_organization_id', true), '')::uuid
$$;

CREATE OR REPLACE FUNCTION identity.resolving_facility() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.resolving_facility_id', true), '')::uuid
$$;

CREATE POLICY internal_resolution ON identity.organizations
    FOR SELECT
    USING (id = identity.resolving_organization());

CREATE POLICY internal_resolution ON identity.facilities
    FOR SELECT
    USING (id = identity.resolving_facility());

CREATE POLICY internal_resolution ON identity.treasury_addresses
    FOR SELECT
    USING (organization_id = identity.resolving_organization());

COMMENT ON FUNCTION identity.resolving_organization() IS
    'Set for the length of one transaction when a service resolves a single organization it names. Possession of the identifier is the grant, so a caller reads exactly the row it asked for and never a listing, which is why this is a policy rather than a role that bypasses row-level security.';
