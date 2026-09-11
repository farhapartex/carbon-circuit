CREATE OR REPLACE FUNCTION credit_ledger.current_organization_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.organization_id', true), '')::uuid
$$;

CREATE OR REPLACE FUNCTION credit_ledger.current_user_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.user_id', true), '')::uuid
$$;

ALTER TABLE credit_ledger.credit_balances ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_ledger.credit_balances FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON credit_ledger.credit_balances
    USING (organization_id = credit_ledger.current_organization_id())
    WITH CHECK (organization_id = credit_ledger.current_organization_id());

ALTER TABLE credit_ledger.credit_issuances ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_ledger.credit_issuances FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON credit_ledger.credit_issuances
    USING (organization_id = credit_ledger.current_organization_id())
    WITH CHECK (organization_id = credit_ledger.current_organization_id());

ALTER TABLE credit_ledger.idempotency_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_ledger.idempotency_records FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON credit_ledger.idempotency_records
    USING (
        organization_id = credit_ledger.current_organization_id()
        OR (organization_id IS NULL AND user_id = credit_ledger.current_user_id())
    )
    WITH CHECK (
        organization_id = credit_ledger.current_organization_id()
        OR (organization_id IS NULL AND user_id = credit_ledger.current_user_id())
    );

COMMENT ON TABLE credit_ledger.credit_classes IS
    'The permanent, inseparable combination of facility, vintage and activity that PRD 2.3 forbids pooling. Deliberately without row-level security: a class says what a credit is, not who holds it, the same class is held by many organizations after a trade, and the public retirement log names all three of its attributes anyway.';
