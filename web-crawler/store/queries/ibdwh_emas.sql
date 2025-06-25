-- name: CreateEmas :one
INSERT INTO ibdwh.emas (emas_id, jual, beli, created_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (emas_id) 
DO UPDATE SET 
    jual = EXCLUDED.jual,
    beli = EXCLUDED.beli,
    created_at = EXCLUDED.created_at
RETURNING *;

-- name: GetAllEmas :many
SELECT * FROM ibdwh.emas
ORDER BY emas_id DESC
LIMIT $1
OFFSET $2;

-- name: GetTotalEmas :one
SELECT COUNT(*) FROM ibdwh.emas;