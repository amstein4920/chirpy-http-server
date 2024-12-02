-- name: DelChirp :exec

delete from chirps where id = $1;