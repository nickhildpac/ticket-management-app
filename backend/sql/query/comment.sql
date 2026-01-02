-- name: CreateComment :one
INSERT INTO comments (description, ticket_id, created_by, updated_at ) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetComment :one
SELECT * FROM comments WHERE id = $1 LIMIT 1;

-- name: ListComment :many
SELECT * FROM comments WHERE ticket_id=$1 ORDER BY created_at LIMIT $2 OFFSET $3;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;