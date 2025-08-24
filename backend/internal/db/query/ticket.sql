-- name: CreateTicket :one
INSERT INTO tickets (title, description, created_by ) VALUES ($1, $2, $3) RETURNING *;

-- name: GetTicket :one
SELECT * FROM tickets WHERE id = $1 LIMIT 1;

-- name: ListTickets :many
SELECT * FROM tickets WHERE created_by=$1 ORDER BY id LIMIT $2 OFFSET $3;

-- name: DeleteTicket :exec
DELETE FROM tickets WHERE id = $1;