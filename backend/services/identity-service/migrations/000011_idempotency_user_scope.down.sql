DROP POLICY tenant_isolation ON identity.idempotency_records;

CREATE POLICY tenant_isolation ON identity.idempotency_records
  USING (organization_id = identity.current_organization_id())
  WITH CHECK (organization_id = identity.current_organization_id());

DROP INDEX identity.idempotency_records_user_id;
DROP INDEX identity.idempotency_records_user_key;
DROP INDEX identity.idempotency_records_tenant_key;

CREATE UNIQUE INDEX idempotency_records_key
  ON identity.idempotency_records (organization_id, endpoint, idempotency_key);

DELETE FROM identity.idempotency_records WHERE organization_id IS NULL;

ALTER TABLE identity.idempotency_records
  DROP CONSTRAINT idempotency_records_scope;

ALTER TABLE identity.idempotency_records
  DROP COLUMN user_id,
  ALTER COLUMN organization_id SET NOT NULL;
