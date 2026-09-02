CREATE OR REPLACE FUNCTION billing.current_user_id()
RETURNS uuid
LANGUAGE sql
STABLE
AS $$
    SELECT nullif(current_setting('app.user_id', true), '')::uuid
$$;

ALTER TABLE billing.idempotency_records
  ALTER COLUMN organization_id DROP NOT NULL,
  ADD COLUMN user_id uuid;

ALTER TABLE billing.idempotency_records
  ADD CONSTRAINT idempotency_records_scope
  CHECK (organization_id IS NOT NULL OR user_id IS NOT NULL);

DROP INDEX billing.idempotency_records_key;

CREATE UNIQUE INDEX idempotency_records_tenant_key
  ON billing.idempotency_records (organization_id, endpoint, idempotency_key)
  WHERE organization_id IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX idempotency_records_user_key
  ON billing.idempotency_records (user_id, endpoint, idempotency_key)
  WHERE organization_id IS NULL AND deleted_at IS NULL;

CREATE INDEX idempotency_records_user_id
  ON billing.idempotency_records (user_id);

DROP POLICY tenant_isolation ON billing.idempotency_records;

CREATE POLICY tenant_isolation ON billing.idempotency_records
  USING (
    organization_id = billing.current_organization_id()
    OR (organization_id IS NULL AND user_id = billing.current_user_id())
  )
  WITH CHECK (
    organization_id = billing.current_organization_id()
    OR (organization_id IS NULL AND user_id = billing.current_user_id())
  );
