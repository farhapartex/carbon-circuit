CREATE TYPE sustainability.ai_assessment AS ENUM (
    'not_assessed',
    'corroborated',
    'corroborated_with_discrepancy',
    'uncorroborated',
    'contradicted'
);

CREATE TABLE sustainability.claim_ai_reviews (
    id                uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz,
    version           integer NOT NULL DEFAULT 1,
    organization_id   uuid NOT NULL,
    claim_id          uuid NOT NULL REFERENCES sustainability.claims (id),
    assessment        sustainability.ai_assessment NOT NULL,
    confidence        numeric(4,3),
    extracted_figures jsonb NOT NULL DEFAULT '{}'::jsonb,
    flags             jsonb NOT NULL DEFAULT '[]'::jsonb,
    citations         jsonb NOT NULL DEFAULT '[]'::jsonb,
    narrative         text NOT NULL DEFAULT '',
    assessed_by       text NOT NULL,
    assessed_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT claim_ai_reviews_confidence_range
        CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1)),
    CONSTRAINT claim_ai_reviews_unassessed_has_no_confidence
        CHECK (assessment <> 'not_assessed' OR confidence IS NULL)
);

CREATE UNIQUE INDEX claim_ai_reviews_claim
    ON sustainability.claim_ai_reviews (claim_id)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN sustainability.claim_ai_reviews.assessed_by IS
    'Which reviewer produced this assessment. A stub standing in for the AI Agent Service records itself here, so a claim reviewed without a real assessment is auditable rather than indistinguishable from one that was.';

ALTER TABLE sustainability.claim_ai_reviews ENABLE ROW LEVEL SECURITY;
ALTER TABLE sustainability.claim_ai_reviews FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sustainability.claim_ai_reviews
    USING (organization_id = sustainability.current_organization_id())
    WITH CHECK (organization_id = sustainability.current_organization_id());
