CREATE TABLE billing.usage_counters (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    organization_id  uuid NOT NULL,
    dimension        billing.usage_dimension NOT NULL,
    period_start     timestamptz NOT NULL,
    period_end       timestamptz NOT NULL,
    used             bigint NOT NULL DEFAULT 0,
    CONSTRAINT usage_counters_period_ordered CHECK (period_end > period_start),
    CONSTRAINT usage_counters_never_negative CHECK (used >= 0)
);

CREATE UNIQUE INDEX usage_counters_unique
    ON billing.usage_counters (organization_id, dimension, period_start) WHERE deleted_at IS NULL;
CREATE INDEX usage_counters_organization_period
    ON billing.usage_counters (organization_id, period_start DESC);
