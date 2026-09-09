ALTER TABLE sustainability.claims
    ADD COLUMN vintage_ceiling numeric(28,6) NOT NULL DEFAULT 0,
    ADD COLUMN consumed_at_submission numeric(28,6) NOT NULL DEFAULT 0;

ALTER TABLE sustainability.claims
    ADD CONSTRAINT claims_ceiling_within_vintage CHECK (
        vintage_ceiling = 0 OR computed_ceiling <= vintage_ceiling
    );

COMMENT ON COLUMN sustainability.claims.vintage_ceiling IS
    'The full-year ceiling for this facility, activity and vintage. A claim period earns a pro-rated share of it, and every claim in the vintage draws from the same figure.';

COMMENT ON COLUMN sustainability.claims.consumed_at_submission IS
    'What earlier claims in the same vintage had already reserved when this one was submitted, recorded so the ceiling on this claim can be explained years later.';
