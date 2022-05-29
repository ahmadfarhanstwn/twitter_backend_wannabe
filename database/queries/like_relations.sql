-- name: CreateLikeRelation :one
INSERT INTO like_relations
(username, tweet_id)
VALUES ($1,$2)
RETURNING *;

-- name: DeleteLikeRelation :exec
DELETE FROM like_relations
WHERE username = $1 AND tweet_id = $2;

-- name: GetLikeRelation :one
SELECT * FROM like_relations
WHERE username = $1 AND tweet_id = $2;