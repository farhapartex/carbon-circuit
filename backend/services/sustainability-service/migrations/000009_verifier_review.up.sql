CREATE OR REPLACE FUNCTION sustainability.current_platform_role() RETURNS text
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.platform_role', true), '')
$$;

CREATE TABLE sustainability.verifier_exclusions (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    verifier_user_id uuid NOT NULL,
    organization_id  uuid NOT NULL,
    declared_by      uuid NOT NULL,
    reason           text NOT NULL
);

CREATE UNIQUE INDEX verifier_exclusions_pair
    ON sustainability.verifier_exclusions (verifier_user_id, organization_id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE sustainability.verifier_exclusions IS
    'Verifier to organization relationships declared by an Admin, per PRD 3.4. Deliberately carries no row-level security: the review policy reads it from inside a policy expression, and a tenant-scoped table would return no rows there, silently disabling every exclusion.';

CREATE OR REPLACE FUNCTION sustainability.reviewable_by_current_verifier(subject uuid)
RETURNS boolean
LANGUAGE sql STABLE AS $$
    SELECT sustainability.current_platform_role() = 'verifier'
       AND NOT EXISTS (
            SELECT 1
            FROM sustainability.verifier_exclusions exclusion
            WHERE exclusion.verifier_user_id = sustainability.current_user_id()
              AND exclusion.organization_id = subject
              AND exclusion.deleted_at IS NULL
       )
$$;

COMMENT ON FUNCTION sustainability.reviewable_by_current_verifier(uuid) IS
    'True only when the caller holds the verifier platform role for this transaction and has not been declared related to the owning organization. The role is set from the verified service token and only on verifier request paths, so an ordinary tenant read never takes this branch.';

CREATE TYPE sustainability.decision_outcome AS ENUM (
    'approved', 'rejected', 'more_information_requested'
);

CREATE TABLE sustainability.claim_decisions (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    organization_id  uuid NOT NULL,
    claim_id         uuid NOT NULL REFERENCES sustainability.claims (id),
    verifier_user_id uuid NOT NULL,
    verifier_name    text NOT NULL,
    outcome          sustainability.decision_outcome NOT NULL,
    approved_amount  numeric(28,6),
    reason           text NOT NULL,
    decided_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT claim_decisions_amount_only_when_approved CHECK (
        (outcome = 'approved') = (approved_amount IS NOT NULL)
    ),
    CONSTRAINT claim_decisions_amount_positive CHECK (
        approved_amount IS NULL OR approved_amount > 0
    ),
    CONSTRAINT claim_decisions_reason_substantive CHECK (
        outcome = 'approved' OR length(btrim(reason)) >= 40
    )
);

CREATE UNIQUE INDEX claim_decisions_one_per_verifier
    ON sustainability.claim_decisions (claim_id, verifier_user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX claim_decisions_claim ON sustainability.claim_decisions (claim_id);

COMMENT ON TABLE sustainability.claim_decisions IS
    'Append only. PRD 3.4 makes a recorded decision final; a wrong one is corrected by a separate Admin-initiated reversal, never by editing this row.';

ALTER TABLE sustainability.claim_decisions ENABLE ROW LEVEL SECURITY;
ALTER TABLE sustainability.claim_decisions FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON sustainability.claim_decisions
    USING (organization_id = sustainability.current_organization_id())
    WITH CHECK (organization_id = sustainability.current_organization_id());

CREATE POLICY verifier_review ON sustainability.claim_decisions
    FOR SELECT
    USING (sustainability.reviewable_by_current_verifier(organization_id));

CREATE POLICY verifier_records ON sustainability.claim_decisions
    FOR INSERT
    WITH CHECK (
        sustainability.reviewable_by_current_verifier(organization_id)
        AND verifier_user_id = sustainability.current_user_id()
    );

CREATE POLICY verifier_review ON sustainability.claims
    FOR SELECT
    USING (sustainability.reviewable_by_current_verifier(organization_id));

CREATE POLICY verifier_review ON sustainability.claim_evidence
    FOR SELECT
    USING (sustainability.reviewable_by_current_verifier(organization_id));

CREATE POLICY verifier_review ON sustainability.claim_ai_reviews
    FOR SELECT
    USING (sustainability.reviewable_by_current_verifier(organization_id));
