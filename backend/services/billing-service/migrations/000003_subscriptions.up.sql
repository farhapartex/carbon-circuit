CREATE TABLE billing.subscriptions (
    id                      uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    deleted_at              timestamptz,
    version                 integer NOT NULL DEFAULT 1,
    organization_id         uuid NOT NULL,
    plan_id                 uuid NOT NULL REFERENCES billing.plans (id),
    state                   billing.subscription_state NOT NULL DEFAULT 'active',
    stripe_customer_id      text,
    stripe_subscription_id  text,
    current_period_start    timestamptz NOT NULL,
    current_period_end      timestamptz NOT NULL,
    grace_period_ends_at    timestamptz,
    cancel_at               timestamptz,
    cancelled_at            timestamptz,
    CONSTRAINT subscriptions_period_ordered CHECK (current_period_end > current_period_start),
    CONSTRAINT subscriptions_grace_only_when_failing
        CHECK (grace_period_ends_at IS NULL OR state IN ('grace_period', 'read_only'))
);

CREATE UNIQUE INDEX subscriptions_one_per_organization
    ON billing.subscriptions (organization_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX subscriptions_stripe_subscription_id
    ON billing.subscriptions (stripe_subscription_id) WHERE stripe_subscription_id IS NOT NULL;
CREATE INDEX subscriptions_plan_id ON billing.subscriptions (plan_id);
CREATE INDEX subscriptions_stripe_customer_id ON billing.subscriptions (stripe_customer_id);
CREATE INDEX subscriptions_grace_expiry ON billing.subscriptions (grace_period_ends_at)
    WHERE state = 'grace_period';

CREATE TABLE billing.plan_overrides (
    id                   uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz,
    version              integer NOT NULL DEFAULT 1,
    organization_id      uuid NOT NULL,
    dimension            billing.usage_dimension NOT NULL,
    included             bigint,
    overage_rate_usd     numeric(10,2),
    justification        text NOT NULL,
    created_by_admin_id  uuid NOT NULL,
    expires_at           timestamptz
);

CREATE UNIQUE INDEX plan_overrides_unique
    ON billing.plan_overrides (organization_id, dimension) WHERE deleted_at IS NULL;
CREATE INDEX plan_overrides_organization_id ON billing.plan_overrides (organization_id);
CREATE INDEX plan_overrides_expiry ON billing.plan_overrides (expires_at) WHERE expires_at IS NOT NULL;
