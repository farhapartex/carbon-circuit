DROP INDEX IF EXISTS identity.sessions_auth0_session;

CREATE UNIQUE INDEX sessions_auth0_session
    ON identity.sessions (user_id, auth0_session_id)
    WHERE deleted_at IS NULL;
