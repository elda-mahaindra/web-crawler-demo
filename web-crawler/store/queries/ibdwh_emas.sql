-- name: CreateEmas :one
INSERT INTO ibdwh.emas (jual, beli, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAllEmas :many
SELECT * FROM ibdwh.emas
ORDER BY created_at ASC
LIMIT $1
OFFSET $2;

-- name: GetTotalEmas :one
SELECT COUNT(*) FROM ibdwh.emas;