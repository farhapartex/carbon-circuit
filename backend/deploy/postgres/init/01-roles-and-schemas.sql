CREATE EXTENSION IF NOT EXISTS citext;

CREATE ROLE identity_service LOGIN PASSWORD 'identity_service';
CREATE ROLE billing_service LOGIN PASSWORD 'billing_service';
CREATE ROLE provenance_service LOGIN PASSWORD 'provenance_service';
CREATE ROLE provenance_read_service LOGIN PASSWORD 'provenance_read_service';
CREATE ROLE evidence_service LOGIN PASSWORD 'evidence_service';
CREATE ROLE sustainability_service LOGIN PASSWORD 'sustainability_service';

CREATE SCHEMA IF NOT EXISTS identity AUTHORIZATION identity_service;
CREATE SCHEMA IF NOT EXISTS billing AUTHORIZATION billing_service;
CREATE SCHEMA IF NOT EXISTS provenance AUTHORIZATION provenance_service;
CREATE SCHEMA IF NOT EXISTS provenance_read AUTHORIZATION provenance_read_service;
CREATE SCHEMA IF NOT EXISTS evidence AUTHORIZATION evidence_service;
CREATE SCHEMA IF NOT EXISTS sustainability AUTHORIZATION sustainability_service;

REVOKE ALL ON SCHEMA public FROM PUBLIC;

GRANT USAGE, CREATE ON SCHEMA identity TO identity_service;
GRANT USAGE, CREATE ON SCHEMA billing TO billing_service;
GRANT USAGE, CREATE ON SCHEMA provenance TO provenance_service;
GRANT USAGE, CREATE ON SCHEMA provenance_read TO provenance_read_service;
GRANT USAGE, CREATE ON SCHEMA evidence TO evidence_service;
GRANT USAGE, CREATE ON SCHEMA sustainability TO sustainability_service;

REVOKE ALL ON SCHEMA billing FROM sustainability_service, identity_service, provenance_service, evidence_service;
REVOKE ALL ON SCHEMA identity FROM sustainability_service, billing_service, provenance_service, evidence_service;
REVOKE ALL ON SCHEMA provenance FROM sustainability_service, identity_service, billing_service, provenance_read_service, evidence_service;
REVOKE ALL ON SCHEMA provenance_read FROM sustainability_service, identity_service, billing_service, provenance_service, evidence_service;

REVOKE ALL ON SCHEMA evidence FROM identity_service, billing_service, provenance_service,
    provenance_read_service, sustainability_service;

REVOKE ALL ON SCHEMA sustainability FROM identity_service, billing_service, provenance_service,
    provenance_read_service, evidence_service;

GRANT USAGE ON SCHEMA public
    TO identity_service, billing_service, provenance_service,
       provenance_read_service, evidence_service, sustainability_service;

ALTER ROLE identity_service SET search_path = identity, public;
ALTER ROLE billing_service SET search_path = billing, public;
ALTER ROLE provenance_service SET search_path = provenance, public;
ALTER ROLE provenance_read_service SET search_path = provenance_read, public;
ALTER ROLE evidence_service SET search_path = evidence, public;
ALTER ROLE sustainability_service SET search_path = sustainability, public;
