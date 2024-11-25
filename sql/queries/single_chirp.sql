-- name: SingleChirp :one
select * from chirps where id = $1;