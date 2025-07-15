-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (@id, @created_at, @updated_at, @name)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE name = @name;

-- name: GetUsers :many
SELECT * FROM users;

-- name: DeleteUsers :exec
DELETE FROM users;