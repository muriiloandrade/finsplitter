-- name: GetCardBrand :one
SELECT * FROM card_brand
WHERE id = $1 LIMIT 1;

-- name: ListCardBrands :many
SELECT * FROM card_brand
WHERE
  ((sqlc.narg(name)::text IS NULL OR sqlc.narg(name)::text = '') OR name ILIKE '%' || sqlc.narg(name)::text || '%')
  AND
  (sqlc.narg(id)::uuid IS NULL OR id = sqlc.narg(id)::uuid)
ORDER BY name;

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
