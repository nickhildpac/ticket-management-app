-- name: CreateUser :one
INSERT INTO users (
    hashed_password,
    first_name,
    last_name,
    email,
    updated_at,
    skills
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUsersByIDs :many
SELECT * FROM users
WHERE id = ANY($1::uuid[]);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAllUsers :many
SELECT id, first_name, last_name, email FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET email = $2, first_name = $3, last_name = $4, role = $5, updated_at = $6, skills = $7
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetAllAgents :many
SELECT * FROM users
WHERE role = 'agent'
ORDER BY created_at;

-- name: GetAutoAssignmentCandidates :many
WITH active_workload AS (
    SELECT
        assignee_id,
        COUNT(*)::int4 AS active_ticket_count
    FROM tickets
    CROSS JOIN LATERAL unnest(COALESCE(assigned_to, ARRAY[]::uuid[])) AS assignee_id
    WHERE state = ANY(sqlc.arg('active_states')::int[])
    GROUP BY assignee_id
)
SELECT
    users.id,
    users.hashed_password,
    users.first_name,
    users.last_name,
    users.email,
    users.role,
    users.updated_at,
    users.created_at,
    users.skills,
    COALESCE(active_workload.active_ticket_count, 0)::int4 AS active_ticket_count
FROM users
LEFT JOIN active_workload ON active_workload.assignee_id = users.id
WHERE users.role = 'agent'
  AND users.skills && sqlc.arg('required_skills')::text[]
ORDER BY active_ticket_count ASC, users.created_at ASC;
