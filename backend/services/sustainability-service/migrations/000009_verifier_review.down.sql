DROP POLICY IF EXISTS verifier_review ON sustainability.claim_ai_reviews;
DROP POLICY IF EXISTS verifier_review ON sustainability.claim_evidence;
DROP POLICY IF EXISTS verifier_review ON sustainability.claims;
DROP TABLE IF EXISTS sustainability.claim_decisions;
DROP TYPE IF EXISTS sustainability.decision_outcome;
DROP FUNCTION IF EXISTS sustainability.reviewable_by_current_verifier(uuid);
DROP TABLE IF EXISTS sustainability.verifier_exclusions;
DROP FUNCTION IF EXISTS sustainability.current_platform_role();
