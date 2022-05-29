-- name: CreateTweet :one
INSERT INTO tweets
(tweet, username)
VALUES ($1,$2)
RETURNING *;

-- name: GetTweet :one
SELECT * FROM tweets
WHERE id = $1 LIMIT 1;

-- name: IncrementLike :one
UPDATE tweets SET
likes = likes + 1
WHERE id = $1
RETURNING *;

-- name: DecrementLike :one
UPDATE tweets SET
likes = likes - 1
WHERE id = $1
RETURNING *;

-- name: GetListTweets :many
SELECT * FROM tweets
WHERE username = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: DeleteTweet :exec
DELETE FROM tweets
WHERE id = $1;