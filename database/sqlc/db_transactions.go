package database

import (
	"context"
	"database/sql"
	"fmt"
)

type Transaction interface {
	Querier
	FollowTx(c context.Context, arg FollowInputArgs) (FollowInputResult,error)
	UnfollowTx(c context.Context, arg FollowInputArgs) error
	LikeTweetTx(c context.Context, arg CreateLikeRelationParams) error
	UnlikeTweetTx(c context.Context, arg DeleteLikeRelationParams) error
}

type DBTransaction struct {
	*Queries
	db *sql.DB
}

func NewTransaction(db *sql.DB) Transaction {
	return &DBTransaction{
		db: db,
		Queries: New(db),
	}
}

func (dbt *DBTransaction) execTransaction(ctx context.Context, fn func(*Queries) error) error {
	tx, err := dbt.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			return fmt.Errorf("transaction error : %v, rollback errorL %v", err, rollbackError)
		}
		return err
	}

	return tx.Commit()
}