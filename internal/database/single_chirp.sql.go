// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: single_chirp.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const singleChirp = `-- name: SingleChirp :one
select id, created_at, updated_at, body, user_id from chirps where id = $1
`

func (q *Queries) SingleChirp(ctx context.Context, id uuid.UUID) (Chirp, error) {
	row := q.db.QueryRowContext(ctx, singleChirp, id)
	var i Chirp
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Body,
		&i.UserID,
	)
	return i, err
}