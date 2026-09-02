ALTER TABLE billing.idempotency_records
  ALTER COLUMN response_body TYPE jsonb USING response_body::text::jsonb;
