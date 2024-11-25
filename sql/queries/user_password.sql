-- name: UserPassword :one
SELECT * FROM users where email = $1;