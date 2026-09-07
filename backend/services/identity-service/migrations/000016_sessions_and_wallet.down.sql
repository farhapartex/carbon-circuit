DROP INDEX IF EXISTS identity.sessions_auth0_session;
ALTER TABLE identity.sessions DROP COLUMN IF EXISTS auth0_session_id;
DROP INDEX IF EXISTS identity.users_personal_wallet;
ALTER TABLE identity.users DROP COLUMN IF EXISTS personal_wallet_address;
