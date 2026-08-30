CREATE TABLE identity.facilities (
    id                       uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now(),
    deleted_at               timestamptz,
    version                  integer NOT NULL DEFAULT 1,
    organization_id          uuid NOT NULL REFERENCES identity.organizations (id),
    name                     text NOT NULL,
    address                  text NOT NULL,
    country_code             char(2) NOT NULL,
    grid_region              identity.grid_region NOT NULL,
    type                     identity.facility_type NOT NULL,
    facility_reference       text,
    verification_status      identity.facility_verification NOT NULL DEFAULT 'self_declared',
    ceiling_discount_factor  numeric(3,2) NOT NULL,
    trust_tier               identity.trust_tier NOT NULL DEFAULT 'new',
    declared_capacity        numeric(20,6) NOT NULL,
    declared_energy_kwh      numeric(20,6) NOT NULL,
    attested_capacity        numeric(20,6),
    attested_energy_kwh      numeric(20,6),
    CONSTRAINT facilities_discount_matches_verification CHECK (
        (verification_status = 'facility_matched' AND ceiling_discount_factor = 1.00)
        OR (verification_status = 'organization_matched' AND ceiling_discount_factor = 0.75)
        OR (verification_status = 'self_declared' AND ceiling_discount_factor = 0.50)
    ),
    CONSTRAINT facilities_declared_figures_positive
        CHECK (declared_capacity > 0 AND declared_energy_kwh > 0)
);

CREATE INDEX facilities_organization_id ON identity.facilities (organization_id);
CREATE INDEX facilities_organization_created
    ON identity.facilities (organization_id, created_at DESC)
    WHERE deleted_at IS NULL;
