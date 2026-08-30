CREATE TABLE identity.business_registry_records (
    id                    uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz,
    version               integer NOT NULL DEFAULT 1,
    country_code          char(2) NOT NULL,
    registration_number   text NOT NULL,
    legal_name            text NOT NULL,
    registered_address    text NOT NULL,
    incorporation_date    date NOT NULL,
    entity_status         identity.registry_entity_status NOT NULL,
    industry_codes        text[] NOT NULL DEFAULT '{}',
    sanctioned            boolean NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX business_registry_records_lookup
    ON identity.business_registry_records (country_code, registration_number)
    WHERE deleted_at IS NULL;

CREATE TABLE identity.facility_registry_records (
    id                                uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at                        timestamptz NOT NULL DEFAULT now(),
    updated_at                        timestamptz NOT NULL DEFAULT now(),
    deleted_at                        timestamptz,
    version                           integer NOT NULL DEFAULT 1,
    organization_registration_number  text NOT NULL,
    facility_reference                text NOT NULL,
    attested_capacity                 numeric(20,6) NOT NULL,
    attested_energy_kwh               numeric(20,6) NOT NULL
);

CREATE UNIQUE INDEX facility_registry_records_lookup
    ON identity.facility_registry_records (organization_registration_number, facility_reference)
    WHERE deleted_at IS NULL;
