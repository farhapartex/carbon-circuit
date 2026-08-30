CREATE OR REPLACE FUNCTION billing.current_organization_id() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.organization_id', true), '')::uuid
$$;

ALTER TABLE billing.subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing.subscriptions FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON billing.subscriptions
    USING (organization_id = billing.current_organization_id())
    WITH CHECK (organization_id = billing.current_organization_id());

ALTER TABLE billing.plan_overrides ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing.plan_overrides FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON billing.plan_overrides
    USING (organization_id = billing.current_organization_id())
    WITH CHECK (organization_id = billing.current_organization_id());

ALTER TABLE billing.usage_counters ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing.usage_counters FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON billing.usage_counters
    USING (organization_id = billing.current_organization_id())
    WITH CHECK (organization_id = billing.current_organization_id());

ALTER TABLE billing.invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing.invoices FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON billing.invoices
    USING (organization_id = billing.current_organization_id())
    WITH CHECK (organization_id = billing.current_organization_id());

ALTER TABLE billing.payment_methods ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing.payment_methods FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON billing.payment_methods
    USING (organization_id = billing.current_organization_id())
    WITH CHECK (organization_id = billing.current_organization_id());

ALTER TABLE billing.idempotency_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing.idempotency_records FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON billing.idempotency_records
    USING (organization_id = billing.current_organization_id())
    WITH CHECK (organization_id = billing.current_organization_id());
