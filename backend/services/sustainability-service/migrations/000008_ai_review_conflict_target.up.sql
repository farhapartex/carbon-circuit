DROP INDEX sustainability.claim_ai_reviews_claim;

CREATE UNIQUE INDEX claim_ai_reviews_claim
    ON sustainability.claim_ai_reviews (claim_id);

COMMENT ON INDEX sustainability.claim_ai_reviews_claim IS
    'Deliberately not partial. Postgres only matches an ON CONFLICT target to a partial index when the statement repeats its predicate, and a review is superseded rather than soft-deleted, so the partial form bought nothing and broke idempotent recording.';
