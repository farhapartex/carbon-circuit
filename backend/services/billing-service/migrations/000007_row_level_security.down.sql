DROP POLICY IF EXISTS tenant_isolation ON billing.idempotency_records;
DROP POLICY IF EXISTS tenant_isolation ON billing.payment_methods;
DROP POLICY IF EXISTS tenant_isolation ON billing.invoices;
DROP POLICY IF EXISTS tenant_isolation ON billing.usage_counters;
DROP POLICY IF EXISTS tenant_isolation ON billing.plan_overrides;
DROP POLICY IF EXISTS tenant_isolation ON billing.subscriptions;

ALTER TABLE billing.idempotency_records DISABLE ROW LEVEL SECURITY;
ALTER TABLE billing.payment_methods DISABLE ROW LEVEL SECURITY;
ALTER TABLE billing.invoices DISABLE ROW LEVEL SECURITY;
ALTER TABLE billing.usage_counters DISABLE ROW LEVEL SECURITY;
ALTER TABLE billing.plan_overrides DISABLE ROW LEVEL SECURITY;
ALTER TABLE billing.subscriptions DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS billing.current_organization_id();
