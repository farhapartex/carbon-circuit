CREATE UNIQUE INDEX treasury_addresses_one_organization_per_address
  ON identity.treasury_addresses (address)
  WHERE state = 'active' AND deleted_at IS NULL;
