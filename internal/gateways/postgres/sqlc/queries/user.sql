-- name: CreateUser :one
INSERT INTO "user" (
    logto_user_id, username, email, created_date, last_modified_date
) VALUES (
    $1, $2, $3, NOW(), NOW()
)
RETURNING id, name, email, phone_number, username, password_hash, logto_user_id, created_date, last_modified_date;

-- name: GetUserByID :one
SELECT id, name, email, phone_number, username, password_hash, logto_user_id, created_date, last_modified_date
FROM "user"
WHERE id = $1;

-- name: GetUserByLogtoUserID :one
SELECT id, name, email, phone_number, username, password_hash, logto_user_id, created_date, last_modified_date
FROM "user"
WHERE logto_user_id = $1;

-- name: UpdateUsername :exec
UPDATE "user"
SET username = $2, last_modified_date = NOW()
WHERE id = $1;

-- name: ExistsByLogtoUserID :one
SELECT EXISTS(SELECT 1 FROM "user" WHERE logto_user_id = $1);
