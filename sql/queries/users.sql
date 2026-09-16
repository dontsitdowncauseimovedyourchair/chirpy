-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
           $1,
        $2,
        $3,
        $4,
        $5
       )
    RETURNING *;

-- name: WipeUsers :exec
DELETE FROM users;

-- name: GetUserByID :one
SELECT * FROM users
WHERE ID = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateEmailPassword :exec
UPDATE users SET email = $2, hashed_password = $3, updated_at = now()
WHERE id = $1;