ALTER TABLE identity.users
    ADD COLUMN personal_wallet_address char(42);

CREATE UNIQUE INDEX users_personal_wallet
    ON identity.users (personal_wallet_address)
    WHERE personal_wallet_address IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE identity.sessions
    ADD COLUMN auth0_session_id text;

UPDATE identity.sessions SET auth0_session_id = id::text WHERE auth0_session_id IS NULL;

ALTER TABLE identity.sessions
    ALTER COLUMN auth0_session_id SET NOT NULL;

CREATE UNIQUE INDEX sessions_auth0_session
    ON identity.sessions (user_id, auth0_session_id)
    WHERE deleted_at IS NULL;
