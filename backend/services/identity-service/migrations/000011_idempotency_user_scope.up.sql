ALTER TABLE identity.idempotency_records
  ALTER COLUMN organization_id DROP NOT NULL,
  ADD COLUMN user_id uuid REFERENCES identity.users(id);

ALTER TABLE identity.idempotency_records
  ADD CONSTRAINT idempotency_records_scope
  CHECK (organization_id IS NOT NULL OR user_id IS NOT NULL);

DROP INDEX identity.idempotency_records_key;

CREATE UNIQUE INDEX idempotency_records_tenant_key
  ON identity.idempotency_records (organization_id, endpoint, idempotency_key)
  WHERE organization_id IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX idempotency_records_user_key
  ON identity.idempotency_records (user_id, endpoint, idempotency_key)
  WHERE organization_id IS NULL AND deleted_at IS NULL;

CREATE INDEX idempotency_records_user_id
  ON identity.idempotency_records (user_id);

DROP POLICY tenant_isolation ON identity.idempotency_records;

CREATE POLICY tenant_isolation ON identity.idempotency_records
  USING (
    organization_id = identity.current_organization_id()
    OR (organization_id IS NULL AND user_id = identity.current_user_id())
  )
  WITH CHECK (
    organization_id = identity.current_organization_id()
    OR (organization_id IS NULL AND user_id = identity.current_user_id())
  );
