-- name: CreateTicket :one
INSERT INTO tickets (title, description, created_by, updated_at, skills, priority) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetTicket :one
SELECT * FROM tickets WHERE id = $1 LIMIT 1;

-- name: ListTickets :many
SELECT * FROM tickets WHERE created_by = sqlc.arg('created_by')
ORDER BY
  CASE WHEN sqlc.arg('sort_val')::integer = 1 THEN ticket_number END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 2 THEN ticket_number END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 3 THEN created_at END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 4 THEN created_at END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 0 THEN id END ASC NULLS LAST
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListAllTickets :many
SELECT * FROM tickets
ORDER BY
  CASE WHEN sqlc.arg('sort_val')::integer = 1 THEN ticket_number END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 2 THEN ticket_number END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 3 THEN created_at END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 4 THEN created_at END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 0 THEN id END ASC NULLS LAST
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListAllTicketsByStatePriority :many
-- Optional filters: pass NULL for any sqlc.narg to skip that condition.
-- filter_assignee: tickets where assigned_to contains this user id.
SELECT id, created_by, assigned_to, title, description, state, priority, created_at, updated_at, ticket_number, skills
FROM tickets
WHERE (sqlc.narg('filter_state')::integer IS NULL OR state = sqlc.narg('filter_state')::integer)
  AND (sqlc.narg('filter_priority')::integer IS NULL OR priority = sqlc.narg('filter_priority')::integer)
  AND (sqlc.narg('filter_created_by')::uuid IS NULL OR created_by = sqlc.narg('filter_created_by')::uuid)
  AND (
    sqlc.narg('filter_assignee')::uuid IS NULL
    OR assigned_to @> ARRAY[sqlc.narg('filter_assignee')::uuid]::uuid[]
  )
  AND (sqlc.narg('filter_ticket_number')::bigint IS NULL OR ticket_number = sqlc.narg('filter_ticket_number')::bigint)
ORDER BY
  CASE WHEN sqlc.arg('sort_val')::integer = 1 THEN ticket_number END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 2 THEN ticket_number END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 3 THEN created_at END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 4 THEN created_at END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 0 THEN id END ASC NULLS LAST
LIMIT sqlc.arg('limit_val') OFFSET sqlc.arg('offset_val');

-- name: ListTicketsAssigned :many
SELECT * FROM tickets WHERE assigned_to @> sqlc.arg('assignee_ids')::uuid[]
ORDER BY
  CASE WHEN sqlc.arg('sort_val')::integer = 1 THEN ticket_number END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 2 THEN ticket_number END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 3 THEN created_at END ASC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 4 THEN created_at END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort_val')::integer = 0 THEN id END ASC NULLS LAST
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: DeleteTicket :exec
DELETE FROM tickets WHERE id = $1;

-- name: GetTicketsByCreator :many
SELECT * FROM tickets
WHERE created_by = $1
ORDER BY created_at DESC;

-- name: GetTicketsByAssignee :many
SELECT * FROM tickets
WHERE assigned_to @> $1::uuid[]
ORDER BY created_at DESC;

-- name: UpdateTicket :one
UPDATE tickets
SET
    title = $2,
    description = $3,
    state = $4,
    priority = $5,
    assigned_to = $6,
    updated_at = $7,
    skills = $8
WHERE id = $1
RETURNING *;

-- name: GetTicketByNumber :one
SELECT * FROM tickets WHERE ticket_number = $1 LIMIT 1;

-- name: GetActiveTickets :many
SELECT * FROM tickets
WHERE state IN (1, 2)
ORDER BY created_at;

-- name: CountTicketStatsAll :one
SELECT
  COUNT(*)::int4 AS total,
  COUNT(*) FILTER (WHERE state = 1)::int4 AS open,
  COUNT(*) FILTER (WHERE state = 2)::int4 AS pending,
  COUNT(*) FILTER (WHERE state = 4)::int4 AS resolved,
  COUNT(*) FILTER (WHERE assigned_to @> $1::uuid[])::int4 AS mine
FROM tickets;

-- name: CountTicketStatsByCreator :one
SELECT
  COUNT(*)::int4 AS total,
  COUNT(*) FILTER (WHERE state = 1)::int4 AS open,
  COUNT(*) FILTER (WHERE state = 2)::int4 AS pending,
  COUNT(*) FILTER (WHERE state = 4)::int4 AS resolved,
  COUNT(*) FILTER (WHERE assigned_to @> $2::uuid[])::int4 AS mine
FROM tickets
WHERE created_by = $1;

-- name: CountTicketStatsByAssignee :one
SELECT
  COUNT(*)::int4 AS total,
  COUNT(*) FILTER (WHERE state = 1)::int4 AS open,
  COUNT(*) FILTER (WHERE state = 2)::int4 AS pending,
  COUNT(*) FILTER (WHERE state = 4)::int4 AS resolved,
  COUNT(*) FILTER (WHERE assigned_to @> $2::uuid[])::int4 AS mine
FROM tickets
WHERE assigned_to @> $1::uuid[];
