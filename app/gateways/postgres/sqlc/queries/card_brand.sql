-- name: GetCardBrand :one
SELECT * FROM card_brand
WHERE id = $1 LIMIT 1;

-- name: ListCardBrands :many
SELECT * FROM card_brand
ORDER BY name;

-- name: CreateCardBrand :one
INSERT INTO card_brand (
  name
) VALUES (
  $1
)
RETURNING *;

-- name: DeleteCardBrand :exec
DELETE FROM card_brand
WHERE id = $1;
