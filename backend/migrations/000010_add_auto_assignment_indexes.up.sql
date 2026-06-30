CREATE INDEX IF NOT EXISTS idx_users_agent_skills
ON users USING GIN (skills)
WHERE role = 'agent';

CREATE INDEX IF NOT EXISTS idx_tickets_active_assigned_to
ON tickets USING GIN (assigned_to)
WHERE state IN (1, 2, 3);

CREATE INDEX IF NOT EXISTS idx_tickets_active_state
ON tickets (state)
WHERE state IN (1, 2, 3);
