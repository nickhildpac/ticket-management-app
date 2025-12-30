-- Ticket Management System - Seed Data
-- Password for all users: 'password123'
-- Generated using: bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

TRUNCATE TABLE comments, tickets, users CASCADE;

-- Users (6 users: 1 admin, 2 agents, 3 users)
INSERT INTO users (id, hashed_password, first_name, last_name, email, role, updated_at, created_at) VALUES
('a1111111-1111-4111-8111-111111111111', '$2a$10$nT3ZUoLgWBwrMCHfyc/QbOXTQNi2DMb4SeOw6Hfoxpe5SBgi36NyO', 'Alice', 'Admin', 'alice@admin.com', 'admin', NOW(), NOW()),
('b2222222-2222-4222-8222-222222222222', '$2a$10$nT3ZUoLgWBwrMCHfyc/QbOXTQNi2DMb4SeOw6Hfoxpe5SBgi36NyO', 'Bob', 'Agent', 'bob@agent.com', 'agent', NOW(), NOW()),
('c3333333-3333-4333-8333-333333333333', '$2a$10$nT3ZUoLgWBwrMCHfyc/QbOXTQNi2DMb4SeOw6Hfoxpe5SBgi36NyO', 'Charlie', 'User', 'charlie@user.com', 'user', NOW(), NOW()),
('d4444444-4444-4444-8444-444444444444', '$2a$10$nT3ZUoLgWBwrMCHfyc/QbOXTQNi2DMb4SeOw6Hfoxpe5SBgi36NyO', 'Diana', 'User', 'diana@user.com', 'user', NOW(), NOW()),
('e5555555-5555-4555-8555-555555555555', '$2a$10$nT3ZUoLgWBwrMCHfyc/QbOXTQNi2DMb4SeOw6Hfoxpe5SBgi36NyO', 'Eve', 'Agent', 'eve@agent.com', 'agent', NOW(), NOW()),
('f6666666-6666-4666-8666-666666666666', '$2a$10$nT3ZUoLgWBwrMCHfyc/QbOXTQNi2DMb4SeOw6Hfoxpe5SBgi36NyO', 'Frank', 'User', 'frank@user.com', 'user', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- Tickets
INSERT INTO tickets (id, created_by, assigned_to, title, description, state, priority, created_at, updated_at) VALUES
('01111111-1111-4111-8111-111111111111', 'c3333333-3333-4333-8333-333333333333', ARRAY['b2222222-2222-4222-8222-222222222222']::UUID[], 'Server Down', 'Production server is down!', 1, 1, NOW() - INTERVAL '1 hour', NOW()),
('02222222-2222-4222-8222-222222222222', 'd4444444-4444-4444-8444-444444444444', ARRAY['e5555555-5555-4555-8555-555555555555']::UUID[], 'Login Issue', 'Cannot login to app', 1, 2, NOW() - INTERVAL '2 hours', NOW()),
('03333333-3333-4333-8333-333333333333', 'f6666666-6666-4666-8666-666666666666', ARRAY[]::UUID[], 'Feature Request', 'Need dark mode', 1, 3, NOW() - INTERVAL '1 day', NOW()),
('04444444-4444-4444-8444-444444444444', 'c3333333-3333-4333-8333-333333333333', ARRAY['b2222222-2222-4222-8222-222222222222', 'e5555555-5555-4555-8555-555555555555']::UUID[], 'API Error', 'Getting 500 errors', 2, 2, NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day'),
('05555555-5555-4555-8555-555555555555', 'd4444444-4444-4444-8444-444444444444', ARRAY[]::UUID[], 'Documentation', 'Update the docs', 3, 4, NOW() - INTERVAL '1 week', NOW() - INTERVAL '2 days'),
('06666666-6666-4666-8666-666666666666', 'f6666666-6666-4666-8666-666666666666', ARRAY['b2222222-2222-4222-8222-222222222222']::UUID[], 'Bug Report', 'Button not working', 3, 3, NOW() - INTERVAL '2 weeks', NOW() - INTERVAL '1 week')
ON CONFLICT DO NOTHING;

-- Comments
INSERT INTO comments (id, ticket_id, created_by, description, created_at, updated_at) VALUES
('00111111-1111-4111-8111-111111111111', '01111111-1111-4111-8111-111111111111', 'c3333333-3333-4333-8333-333333333333', 'This is urgent!', NOW() - INTERVAL '1 hour', NOW()),
('00111111-2222-4222-8222-222222222222', '01111111-1111-4111-8111-111111111111', 'b2222222-2222-4222-8222-222222222222', 'Looking into it.', NOW() - INTERVAL '50 minutes', NOW()),
('00111111-3333-4333-8333-333333333333', '01111111-1111-4111-8111-111111111111', 'c3333333-3333-4333-8333-333333333333', 'Thanks!', NOW() - INTERVAL '30 minutes', NOW()),
('00222222-1111-4111-8111-111111111112', '02222222-2222-4222-8222-222222222222', 'd4444444-4444-4444-8444-444444444444', 'Cannot login after password reset', NOW() - INTERVAL '2 hours', NOW()),
('00222222-2222-4222-8222-222222222223', '02222222-2222-4222-8222-222222222222', 'e5555555-5555-4555-8555-555555555555', 'What error do you see?', NOW() - INTERVAL '1 hour 50 minutes', NOW()),
('00222222-3333-4333-8333-333333333334', '02222222-2222-4222-8222-222222222222', 'd4444444-4444-4444-8444-444444444444', 'Invalid username or password', NOW() - INTERVAL '1 hour 45 minutes', NOW())
ON CONFLICT DO NOTHING;
