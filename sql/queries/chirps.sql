-- name: CreateChirp :one
INSERT INTO chirps(id, created_at, updated_at, body, user_id)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
       )
RETURNING *;

-- name: FetchAllChirps :many
SELECT * FROM chirps
ORDER BY
    CASE WHEN @is_desc::boolean = false THEN created_at END ASC,
    CASE WHEN @is_desc::boolean = true THEN created_at END DESC;

-- name: FetchChirpById :one
SELECT * FROM chirps
WHERE id = $1;

-- name: FetchChirpsByAuthorID :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY
    CASE WHEN @is_desc::boolean = false THEN created_at END ASC,
    CASE WHEN @is_desc::boolean = true THEN created_at END DESC;


-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1;
