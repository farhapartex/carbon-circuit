ALTER TABLE sustainability.claims DROP CONSTRAINT claims_ceiling_within_vintage;
ALTER TABLE sustainability.claims
    DROP COLUMN consumed_at_submission,
    DROP COLUMN vintage_ceiling;
