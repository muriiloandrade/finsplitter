-- name: CreateUser :one
INSERT INTO "user" (name, email, phone_number, username, password_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, email, phone_number, username, password_hash, created_date, last_modified_date;

-- name: GetUser :one
SELECT id, name, email, phone_number, username, password_hash, created_date, last_modified_date
FROM "user"
WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email, phone_number, username, password_hash, created_date, last_modified_date
FROM "user"
ORDER BY name;

-- name: UpdateUser :one
UPDATE "user"
SET name = $1, email = $2, phone_number = $3, username = $4, password_hash = $5
WHERE id = $6
RETURNING id, name, email, phone_number, username, password_hash, created_date, last_modified_date;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE id = $1;
