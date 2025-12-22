-- name: CreateTicket :one
INSERT INTO tickets (title, description, created_by ) VALUES ($1, $2, $3) RETURNING *;

-- name: GetTicket :one
SELECT * FROM tickets WHERE id = $1 LIMIT 1;

-- name: ListTickets :many
SELECT * FROM tickets WHERE created_by=$1 ORDER BY id LIMIT $2 OFFSET $3;

-- name: ListAllTickets :many
SELECT * FROM tickets ORDER BY id LIMIT $1 OFFSET $2;

-- name: ListTicketsAssigned :many
SELECT * FROM tickets WHERE assigned_to=$1 ORDER BY id LIMIT $2 OFFSET $3;

-- name: DeleteTicket :exec
DELETE FROM tickets WHERE id = $1;

-- name: GetTicketsByCreator :many
SELECT * FROM tickets
WHERE created_by = $1
ORDER BY created_at DESC;

-- name: GetTicketsByAssignee :many
SELECT * FROM tickets
WHERE assigned_to = $1
ORDER BY created_at DESC;

-- name: UpdateTicket :one
UPDATE tickets
SET 
    title = $2,
    description = $3,
    state = $4,
    priority = $5,
    assigned_to = $6,
    updated_at = $7
WHERE id = $1
RETURNING *;