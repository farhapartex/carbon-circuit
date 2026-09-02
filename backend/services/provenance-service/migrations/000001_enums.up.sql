CREATE TYPE provenance.product_category AS ENUM ('electronics', 'agriculture', 'pharma', 'textiles');

CREATE TYPE provenance.checkpoint_type AS ENUM (
    'production_complete', 'departed_origin', 'customs_export',
    'customs_import', 'arrived_destination'
);

CREATE TYPE provenance.shipping_method AS ENUM (
    'air_freight_short_haul', 'air_freight_long_haul',
    'sea_freight_container', 'sea_freight_bulk',
    'rail_electric', 'rail_diesel',
    'road_hgv', 'road_lgv', 'inland_waterway'
);

CREATE TYPE provenance.anchor_status AS ENUM ('unanchored', 'provisional', 'confirmed');

CREATE TYPE provenance.idempotency_state AS ENUM ('processing', 'completed', 'failed');
