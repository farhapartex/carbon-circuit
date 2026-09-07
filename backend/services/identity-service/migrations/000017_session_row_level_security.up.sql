ALTER TABLE identity.sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE identity.sessions FORCE ROW LEVEL SECURITY;

CREATE POLICY user_isolation ON identity.sessions
    USING (user_id = identity.current_user_id())
    WITH CHECK (user_id = identity.current_user_id());
