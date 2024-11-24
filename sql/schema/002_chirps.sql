-- +goose Up
CREATE TABLE chirps (
    id uuid PRIMARY KEY,
    created_at timestamp not null,
    updated_at timestamp not null,
    body text not null,
    user_id uuid not null,
    FOREIGN KEY (user_id)
    REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;