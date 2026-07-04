-- Dedicated service account the AI/RAG worker authenticates as when calling back
-- into the ticket API (ADR 0002). It is an admin so it can comment on any ticket
-- (see CanCommentOnTicket); comments.created_by has an FK to users, so this row
-- must exist. The password is a non-bcrypt placeholder so the account cannot log in.
INSERT INTO users (id, hashed_password, first_name, last_name, email, role, skills, updated_at, created_at)
VALUES (
  '00000000-0000-4000-8000-0000000000a1',
  '!disabled',
  'AI',
  'Triage',
  'ai-triage@service.local',
  'admin',
  '{}',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;
