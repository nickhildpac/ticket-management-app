-- name: CreateTicket :one
INSERT INTO tickets (title, description, created_by, updated_at, skills) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetTicket :one
SELECT * FROM tickets WHERE id = $1 LIMIT 1;

-- name: ListTickets :many
SELECT * FROM tickets WHERE created_by=$1 ORDER BY id LIMIT $2 OFFSET $3;

-- name: ListAllTickets :many
SELECT * FROM tickets ORDER BY id LIMIT $1 OFFSET $2;

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
ORDER BY id
LIMIT sqlc.arg('limit_val') OFFSET sqlc.arg('offset_val');

-- name: ListTicketsAssigned :many
SELECT * FROM tickets WHERE assigned_to @> $1::uuid[] ORDER BY id LIMIT $2 OFFSET $3;

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
