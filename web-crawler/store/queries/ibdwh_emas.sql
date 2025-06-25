-- name: CreateEmas :one
INSERT INTO ibdwh.emas (jual, beli, created_at)
VALUES ($1, $2, $3)
RETURNING *;