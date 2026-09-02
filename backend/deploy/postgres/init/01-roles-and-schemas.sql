CREATE EXTENSION IF NOT EXISTS citext;

CREATE ROLE identity_service LOGIN PASSWORD 'identity_service';
CREATE ROLE billing_service LOGIN PASSWORD 'billing_service';
CREATE ROLE provenance_service LOGIN PASSWORD 'provenance_service';
CREATE ROLE provenance_read_service LOGIN PASSWORD 'provenance_read_service';

CREATE SCHEMA IF NOT EXISTS identity AUTHORIZATION identity_service;
CREATE SCHEMA IF NOT EXISTS billing AUTHORIZATION billing_service;
CREATE SCHEMA IF NOT EXISTS provenance AUTHORIZATION provenance_service;
CREATE SCHEMA IF NOT EXISTS provenance_read AUTHORIZATION provenance_read_service;

REVOKE ALL ON SCHEMA public FROM PUBLIC;

GRANT USAGE, CREATE ON SCHEMA identity TO identity_service;
GRANT USAGE, CREATE ON SCHEMA billing TO billing_service;
GRANT USAGE, CREATE ON SCHEMA provenance TO provenance_service;
GRANT USAGE, CREATE ON SCHEMA provenance_read TO provenance_read_service;

REVOKE ALL ON SCHEMA billing FROM identity_service, provenance_service;
REVOKE ALL ON SCHEMA identity FROM billing_service, provenance_service;
REVOKE ALL ON SCHEMA provenance FROM identity_service, billing_service, provenance_read_service;
REVOKE ALL ON SCHEMA provenance_read FROM identity_service, billing_service, provenance_service;

GRANT USAGE ON SCHEMA public
    TO identity_service, billing_service, provenance_service, provenance_read_service;

ALTER ROLE identity_service SET search_path = identity, public;
ALTER ROLE billing_service SET search_path = billing, public;
ALTER ROLE provenance_service SET search_path = provenance, public;
ALTER ROLE provenance_read_service SET search_path = provenance_read, public;
