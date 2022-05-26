// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: relations.sql

package database

import (
	"context"
)

const createRelations = `-- name: CreateRelations :one
INSERT INTO relations
(follower_username, followed_username)
VALUES ($1,$2)
RETURNING id, follower_username, followed_username, created_at
`

type CreateRelationsParams struct {
	FollowerUsername string `json:"follower_username"`
	FollowedUsername string `json:"followed_username"`
}

func (q *Queries) CreateRelations(ctx context.Context, arg CreateRelationsParams) (Relations, error) {
	row := q.db.QueryRowContext(ctx, createRelations, arg.FollowerUsername, arg.FollowedUsername)
	var i Relations
	err := row.Scan(
		&i.ID,
		&i.FollowerUsername,
		&i.FollowedUsername,
		&i.CreatedAt,
	)
	return i, err
}

const deleteRelation = `-- name: DeleteRelation :exec
DELETE FROM relations
WHERE follower_username = $1 AND followed_username = $2
`

type DeleteRelationParams struct {
	FollowerUsername string `json:"follower_username"`
	FollowedUsername string `json:"followed_username"`
}

func (q *Queries) DeleteRelation(ctx context.Context, arg DeleteRelationParams) error {
	_, err := q.db.ExecContext(ctx, deleteRelation, arg.FollowerUsername, arg.FollowedUsername)
	return err
}

const getFollower = `-- name: GetFollower :many
SELECT id, follower_username, followed_username, created_at FROM relations
WHERE followed_username = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3
`

type GetFollowerParams struct {
	FollowedUsername string `json:"followed_username"`
	Limit            int32  `json:"limit"`
	Offset           int32  `json:"offset"`
}

func (q *Queries) GetFollower(ctx context.Context, arg GetFollowerParams) ([]Relations, error) {
	rows, err := q.db.QueryContext(ctx, getFollower, arg.FollowedUsername, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Relations
	for rows.Next() {
		var i Relations
		if err := rows.Scan(
			&i.ID,
			&i.FollowerUsername,
			&i.FollowedUsername,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFollowing = `-- name: GetFollowing :many
SELECT id, follower_username, followed_username, created_at FROM relations
WHERE follower_username = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3
`

type GetFollowingParams struct {
	FollowerUsername string `json:"follower_username"`
	Limit            int32  `json:"limit"`
	Offset           int32  `json:"offset"`
}

func (q *Queries) GetFollowing(ctx context.Context, arg GetFollowingParams) ([]Relations, error) {
	rows, err := q.db.QueryContext(ctx, getFollowing, arg.FollowerUsername, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Relations
	for rows.Next() {
		var i Relations
		if err := rows.Scan(
			&i.ID,
			&i.FollowerUsername,
			&i.FollowedUsername,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRelations = `-- name: GetRelations :one
SELECT id, follower_username, followed_username, created_at FROM relations
WHERE follower_username = $1 AND followed_username = $2
LIMIT 1
`

type GetRelationsParams struct {
	FollowerUsername string `json:"follower_username"`
	FollowedUsername string `json:"followed_username"`
}

func (q *Queries) GetRelations(ctx context.Context, arg GetRelationsParams) (Relations, error) {
	row := q.db.QueryRowContext(ctx, getRelations, arg.FollowerUsername, arg.FollowedUsername)
	var i Relations
	err := row.Scan(
		&i.ID,
		&i.FollowerUsername,
		&i.FollowedUsername,
		&i.CreatedAt,
	)
	return i, err
}