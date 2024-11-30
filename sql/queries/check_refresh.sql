-- name: CheckRefresh :one
select user_id from refresh_tokens where token = $1 and expires_at < NOW() and revoked_at is null;

-- name: UpdateRevocation :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() where token = $1;