-- name: AllChirps :many
select * from chirps order by created_at;

-- name: AllChirpsAuthorID :many
select * from chirps where user_id = $1;