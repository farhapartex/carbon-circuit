ALTER TABLE sustainability.claim_evidence
    ADD COLUMN byte_size bigint NOT NULL DEFAULT 0;

ALTER TABLE sustainability.claims
    ADD COLUMN reference_lookup_key text NOT NULL DEFAULT '';

COMMENT ON COLUMN sustainability.claims.reference_lookup_key IS
    'The key the reference factor was looked up by, taken from the facility registration rather than the submission, so the claim records which factor actually applied. Added with a default rather than backfilled by UPDATE, because a migration runs under forced row-level security with no organization context and would see no rows to update.';
