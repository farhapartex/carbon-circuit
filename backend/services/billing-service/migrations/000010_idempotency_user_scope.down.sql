DROP POLICY tenant_isolation ON billing.idempotency_records;

CREATE POLICY tenant_isolation ON billing.idempotency_records
  USING (organization_id = billing.current_organization_id())
  WITH CHECK (organization_id = billing.current_organization_id());

DROP INDEX billing.idempotency_records_user_id;
DROP INDEX billing.idempotency_records_user_key;
DROP INDEX billing.idempotency_records_tenant_key;

CREATE UNIQUE INDEX idempotency_records_key
  ON billing.idempotency_records (organization_id, endpoint, idempotency_key);

DELETE FROM billing.idempotency_records WHERE organization_id IS NULL;

ALTER TABLE billing.idempotency_records
  DROP CONSTRAINT idempotency_records_scope;

ALTER TABLE billing.idempotency_records
  DROP COLUMN user_id,
  ALTER COLUMN organization_id SET NOT NULL;

DROP FUNCTION billing.current_user_id();
