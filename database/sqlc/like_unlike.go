package database

import "context"

func (dbt *DBTransaction) LikeTweetTx(c context.Context, arg CreateLikeRelationParams) error {
	err := dbt.execTransaction(c, func(q *Queries) error {
		_, err := q.CreateLikeRelation(c, arg)
		if err != nil {
			return err
		}

		_, err = q.IncrementLike(c, arg.TweetID)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (dbt *DBTransaction) UnlikeTweetTx(c context.Context, arg DeleteLikeRelationParams) error {
	err := dbt.execTransaction(c, func(q *Queries) error {
		err := q.DeleteLikeRelation(c, arg)
		if err != nil {
			return err
		}

		_, err = q.DecrementLike(c, arg.TweetID)
		if err != nil {
			return err
		}

		return nil
	})
	return err
}