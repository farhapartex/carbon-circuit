CREATE TABLE billing.plans (
    id                          uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at                  timestamptz NOT NULL DEFAULT now(),
    updated_at                  timestamptz NOT NULL DEFAULT now(),
    deleted_at                  timestamptz,
    version                     integer NOT NULL DEFAULT 1,
    tier                        billing.plan_tier NOT NULL,
    name                        text NOT NULL,
    audience                    text NOT NULL,
    monthly_price_usd           numeric(10,2) NOT NULL,
    price_note                  text,
    allowed_organization_types  billing.organization_type[] NOT NULL,
    evidence_storage_gb         integer,
    portal_rate_per_minute      integer NOT NULL,
    api_rate_per_minute         integer,
    api_key_limit               integer,
    marketplace_fee_bps         smallint,
    review_turnaround           text NOT NULL,
    support_level               text NOT NULL,
    effective_from              timestamptz NOT NULL DEFAULT now(),
    effective_to                timestamptz,
    CONSTRAINT plans_price_not_negative CHECK (monthly_price_usd >= 0),
    CONSTRAINT plans_fee_within_contract_bounds
        CHECK (marketplace_fee_bps IS NULL OR marketplace_fee_bps BETWEEN 0 AND 1000),
    CONSTRAINT plans_effective_window_ordered
        CHECK (effective_to IS NULL OR effective_to > effective_from),
    CONSTRAINT plans_allows_at_least_one_type
        CHECK (cardinality(allowed_organization_types) > 0)
);

CREATE UNIQUE INDEX plans_tier_effective ON billing.plans (tier, effective_from) WHERE deleted_at IS NULL;
CREATE INDEX plans_current ON billing.plans (tier) WHERE effective_to IS NULL AND deleted_at IS NULL;

CREATE TABLE billing.plan_limits (
    id                    uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz,
    version               integer NOT NULL DEFAULT 1,
    plan_id               uuid NOT NULL REFERENCES billing.plans (id) ON DELETE CASCADE,
    dimension             billing.usage_dimension NOT NULL,
    included              bigint,
    fair_use_ceiling      bigint,
    overage_rate_usd      numeric(10,2),
    blocks_on_exhaustion  boolean NOT NULL DEFAULT true,
    CONSTRAINT plan_limits_included_or_ceiling
        CHECK (included IS NOT NULL OR fair_use_ceiling IS NOT NULL),
    CONSTRAINT plan_limits_overage_implies_no_block
        CHECK (overage_rate_usd IS NULL OR blocks_on_exhaustion = false)
);

CREATE UNIQUE INDEX plan_limits_unique ON billing.plan_limits (plan_id, dimension) WHERE deleted_at IS NULL;
CREATE INDEX plan_limits_plan_id ON billing.plan_limits (plan_id);
