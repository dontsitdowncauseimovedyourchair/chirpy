-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6
       )
RETURNING *;

-- name: FetchToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: FetchUserIDFromToken :one
SELECT user_id  FROM users u
INNER JOIN refresh_tokens r on u.id = r.user_id
WHERE r.token = $1 AND r.expires_at > now() AND r.revoked_at IS NULL;


-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = now(), updated_at = now()
WHERE token = $1;
