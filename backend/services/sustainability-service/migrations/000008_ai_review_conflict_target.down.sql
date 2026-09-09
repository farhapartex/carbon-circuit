DROP INDEX sustainability.claim_ai_reviews_claim;

CREATE UNIQUE INDEX claim_ai_reviews_claim
    ON sustainability.claim_ai_reviews (claim_id)
    WHERE deleted_at IS NULL;
