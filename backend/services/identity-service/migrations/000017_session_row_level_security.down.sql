DROP POLICY IF EXISTS user_isolation ON identity.sessions;
ALTER TABLE identity.sessions NO FORCE ROW LEVEL SECURITY;
ALTER TABLE identity.sessions DISABLE ROW LEVEL SECURITY;
