ALTER TABLE identity.idempotency_records
  ALTER COLUMN response_body TYPE json USING response_body::text::json;
