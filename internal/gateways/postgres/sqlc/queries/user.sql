-- name: CreateUser :one
INSERT INTO "user" (
    logto_user_id, created_date, last_modified_date
) VALUES (
    $1, NOW(), NOW()
)
RETURNING id, logto_user_id, created_date, last_modified_date;

-- name: GetUserByID :one
SELECT id, logto_user_id, created_date, last_modified_date
FROM "user"
WHERE id = $1;

-- name: GetUserByLogtoUserID :one
SELECT id, logto_user_id, created_date, last_modified_date
FROM "user"
WHERE logto_user_id = $1;

-- name: ExistsByLogtoUserID :one
SELECT EXISTS(SELECT 1 FROM "user" WHERE logto_user_id = $1);
