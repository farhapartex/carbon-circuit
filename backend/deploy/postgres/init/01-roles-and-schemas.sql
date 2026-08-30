CREATE ROLE identity_service LOGIN PASSWORD 'identity_service';
CREATE ROLE billing_service LOGIN PASSWORD 'billing_service';

CREATE SCHEMA IF NOT EXISTS identity AUTHORIZATION identity_service;
CREATE SCHEMA IF NOT EXISTS billing AUTHORIZATION billing_service;

REVOKE ALL ON SCHEMA public FROM PUBLIC;

GRANT USAGE, CREATE ON SCHEMA identity TO identity_service;
GRANT USAGE, CREATE ON SCHEMA billing TO billing_service;

REVOKE ALL ON SCHEMA billing FROM identity_service;
REVOKE ALL ON SCHEMA identity FROM billing_service;

ALTER ROLE identity_service SET search_path = identity;
ALTER ROLE billing_service SET search_path = billing;
