-- name: HashPass :one
SELECT hashed_password from users where email = $1;