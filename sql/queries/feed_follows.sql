-- name: CreateFeedFollow :one
INSERT INTO feed_follows(
  id, created_at, updated_at, feed_id, user_id
) VALUES($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserFeedFollows :many
SELECT * FROM feed_follows
WHERE user_id = $1;
