// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: like_relations.sql

package database

import (
	"context"
)

const createLikeRelation = `-- name: CreateLikeRelation :one
INSERT INTO like_relations
(username, tweet_id)
VALUES ($1,$2)
RETURNING id, username, tweet_id, created_at
`

type CreateLikeRelationParams struct {
	Username string `json:"username"`
	TweetID  int64  `json:"tweet_id"`
}

func (q *Queries) CreateLikeRelation(ctx context.Context, arg CreateLikeRelationParams) (LikeRelations, error) {
	row := q.db.QueryRowContext(ctx, createLikeRelation, arg.Username, arg.TweetID)
	var i LikeRelations
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.TweetID,
		&i.CreatedAt,
	)
	return i, err
}

const deleteLikeRelation = `-- name: DeleteLikeRelation :exec
DELETE FROM like_relations
WHERE username = $1 AND tweet_id = $2
`

type DeleteLikeRelationParams struct {
	Username string `json:"username"`
	TweetID  int64  `json:"tweet_id"`
}

func (q *Queries) DeleteLikeRelation(ctx context.Context, arg DeleteLikeRelationParams) error {
	_, err := q.db.ExecContext(ctx, deleteLikeRelation, arg.Username, arg.TweetID)
	return err
}

const getLikeRelation = `-- name: GetLikeRelation :one
SELECT id, username, tweet_id, created_at FROM like_relations
WHERE username = $1 AND tweet_id = $2
`

type GetLikeRelationParams struct {
	Username string `json:"username"`
	TweetID  int64  `json:"tweet_id"`
}

func (q *Queries) GetLikeRelation(ctx context.Context, arg GetLikeRelationParams) (LikeRelations, error) {
	row := q.db.QueryRowContext(ctx, getLikeRelation, arg.Username, arg.TweetID)
	var i LikeRelations
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.TweetID,
		&i.CreatedAt,
	)
	return i, err
}
