CREATE OR REPLACE FUNCTION identity.validating_prefix() RETURNS text
LANGUAGE sql STABLE AS $$
    SELECT nullif(current_setting('app.validating_api_key_prefix', true), '')
$$;

CREATE POLICY prefix_validation ON identity.api_keys
    FOR SELECT
    USING (prefix = identity.validating_prefix());
