-- name: CreateUser :one
INSERT INTO users
(username, email, hashed_password, name)
VALUES ($1,$2,$3,$4)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users SET 
email = $1,
hashed_password= $2,
name = $3
WHERE username = $4
RETURNING *;

-- name: IncrementFollowing :one
UPDATE users SET
following_count = following_count + 1
WHERE username = $1
RETURNING *;

-- name: DecrementFollowing :one
UPDATE users SET
following_count = following_count - 1
WHERE username = $1
RETURNING *;

-- name: IncrementFollower :one
UPDATE users SET
followers_count = followers_count + 1
WHERE username = $1
RETURNING *;

-- name: DecrementFollower :one
UPDATE users SET
followers_count = followers_count - 1
WHERE username = $1
RETURNING *;