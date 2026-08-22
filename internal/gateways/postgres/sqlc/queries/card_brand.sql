-- name: GetCardBrand :one
SELECT * FROM card_brand
WHERE id = $1 LIMIT 1;

-- name: ListCardBrands :many
SELECT id, name, created_date, last_modified_date
FROM card_brand cb
WHERE
  (sqlc.narg(id)::text IS NULL OR cb.id = sqlc.narg(id)::uuid)
  AND
  (sqlc.narg(name)::text IS NULL OR sqlc.narg(name)::text = '' OR cb.name ILIKE '%' || sqlc.narg(name)::text || '%')
ORDER BY name
LIMIT sqlc.arg(page_size)::bigint OFFSET sqlc.arg(page_offset)::bigint;

-- name: CreateCardBrand :one
INSERT INTO card_brand (
  name
) VALUES (
  $1
)
RETURNING *;

-- name: UpdateCardBrand :one
UPDATE card_brand
SET name = $1
WHERE id = $2
RETURNING *;

-- name: DeleteCardBrand :one
DELETE FROM card_brand
WHERE id = $1
RETURNING *;
