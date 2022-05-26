-- name: CreateRelations :one
INSERT INTO relations
(follower_username, followed_username)
VALUES ($1,$2)
RETURNING *;

-- name: GetRelations :one
SELECT * FROM relations
WHERE follower_username = $1 AND followed_username = $2
LIMIT 1;

-- name: GetFollowing :many
SELECT * FROM relations
WHERE follower_username = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: GetFollower :many
SELECT * FROM relations
WHERE followed_username = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: DeleteRelation :exec
DELETE FROM relations
WHERE follower_username = $1 AND followed_username = $2;