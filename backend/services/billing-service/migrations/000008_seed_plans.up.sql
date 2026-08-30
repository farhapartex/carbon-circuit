INSERT INTO billing.plans (
    tier, name, audience, monthly_price_usd, price_note, allowed_organization_types,
    evidence_storage_gb, portal_rate_per_minute, api_rate_per_minute, api_key_limit,
    marketplace_fee_bps, review_turnaround, support_level
) VALUES
(
    'buyer', 'Buyer', 'Companies that only purchase and retire credits',
    0.00, 'Free', ARRAY['credit_buyer']::billing.organization_type[],
    NULL, 300, NULL, NULL, NULL, 'Not applicable', 'Community'
),
(
    'starter', 'Starter', 'Small facilities or logistics partners just getting started',
    49.00, NULL,
    ARRAY['manufacturer','assembler','logistics','credit_buyer']::billing.organization_type[],
    5, 300, NULL, NULL, 300, 'Standard queue', 'Community and email'
),
(
    'growth', 'Growth', 'Mid-sized manufacturers and assemblers with regular volume',
    199.00, NULL,
    ARRAY['manufacturer','assembler','logistics','credit_buyer']::billing.organization_type[],
    50, 300, 600, 5, 250, 'Standard queue', 'Priority email'
),
(
    'enterprise', 'Enterprise',
    'Large manufacturers, multi-facility organizations, and companies with compliance requirements',
    999.00, 'Custom, from $999 per month, contract-negotiated',
    ARRAY['manufacturer','assembler','logistics','credit_buyer']::billing.organization_type[],
    500, 600, 6000, 25, 150, 'Expedited, 24 business hours', 'Dedicated support contact'
);

INSERT INTO billing.plan_limits (plan_id, dimension, included, fair_use_ceiling, overage_rate_usd, blocks_on_exhaustion)
SELECT p.id, d.dimension::billing.usage_dimension, d.included, d.fair_use_ceiling, d.overage_rate_usd, d.blocks_on_exhaustion
FROM billing.plans p
JOIN (VALUES
    ('buyer',      'users',       5::bigint,       NULL::bigint,     NULL::numeric, true),

    ('starter',    'batches',     50,              NULL,             NULL,          true),
    ('starter',    'checkpoints', 500,             NULL,             NULL,          true),
    ('starter',    'facilities',  1,               NULL,             NULL,          true),
    ('starter',    'users',       5,               NULL,             NULL,          true),
    ('starter',    'ai_reviews',  5,               NULL,             NULL,          true),
    ('starter',    'storage_gb',  5,               NULL,             NULL,          true),

    ('growth',     'batches',     1000,            NULL,             NULL,          true),
    ('growth',     'checkpoints', 20000,           NULL,             NULL,          true),
    ('growth',     'facilities',  5,               NULL,             NULL,          true),
    ('growth',     'users',       25,              NULL,             NULL,          true),
    ('growth',     'ai_reviews',  50,              NULL,             5.00,          false),
    ('growth',     'storage_gb',  50,              NULL,             NULL,          true),

    ('enterprise', 'batches',     NULL,            50000,            NULL,          false),
    ('enterprise', 'checkpoints', NULL,            1000000,          NULL,          false),
    ('enterprise', 'facilities',  NULL,            200,              NULL,          false),
    ('enterprise', 'users',       NULL,            500,              NULL,          false),
    ('enterprise', 'ai_reviews',  500,             NULL,             3.00,          false),
    ('enterprise', 'storage_gb',  500,             NULL,             NULL,          true)
) AS d(tier, dimension, included, fair_use_ceiling, overage_rate_usd, blocks_on_exhaustion)
  ON d.tier::billing.plan_tier = p.tier
WHERE p.effective_to IS NULL;
