-- name: UpdatePassEmail :one
update users set email = $3, hashed_password = $2 where id = $1
returning *;
