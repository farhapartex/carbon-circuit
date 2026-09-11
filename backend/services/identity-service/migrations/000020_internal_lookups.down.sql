DROP POLICY IF EXISTS internal_resolution ON identity.treasury_addresses;
DROP POLICY IF EXISTS internal_resolution ON identity.facilities;
DROP POLICY IF EXISTS internal_resolution ON identity.organizations;
DROP FUNCTION IF EXISTS identity.resolving_facility();
DROP FUNCTION IF EXISTS identity.resolving_organization();
