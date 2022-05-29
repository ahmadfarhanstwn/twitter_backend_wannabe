package database

import "context"

type FollowInputArgs struct {
	Username   string `json:"username"`
	FollowUser string `json:"follow_user"`
}

//simplified, just for test
type FollowInputResult struct {
	FollowerUser           string `json:"follower_user"`
	FollowedUser           string `json:"followed_user"`
	FollowerFollowingCount int32  `json:"follower_following_count"`
	FollowedFollowerCount  int32  `json:"followed_follower_count"`
}

func (dbt *DBTransaction) FollowTx(c context.Context, arg FollowInputArgs) (FollowInputResult, error) {
	var res FollowInputResult

	err := dbt.execTransaction(c, func(q *Queries) error {
		//create relation
		createArg := CreateRelationsParams{
			FollowerUsername: arg.Username,
			FollowedUsername: arg.FollowUser,
		}
		rel, err := q.CreateRelations(c, createArg)
		if err != nil {
			return err
		}

		res.FollowedUser = rel.FollowedUsername
		res.FollowerUser = rel.FollowerUsername

		//increment following
		ifollowing, err := q.IncrementFollowing(c, arg.Username)
		if err != nil {
			return err
		}

		res.FollowerFollowingCount = ifollowing.FollowingCount.Int32

		//increment follower
		ifollower, err := q.IncrementFollower(c, arg.FollowUser)
		if err != nil {
			return err
		}

		res.FollowedFollowerCount = ifollower.FollowersCount.Int32

		return nil
	})

	return res, err
}

func (dbt *DBTransaction) UnfollowTx(c context.Context, arg FollowInputArgs) error {
	// var res FollowInputResult

	err := dbt.execTransaction(c, func(q *Queries) error {
		//delete relation
		deleteArg := DeleteRelationParams{
			FollowerUsername: arg.Username,
			FollowedUsername: arg.FollowUser,
		}
		err := q.DeleteRelation(c, deleteArg)
		if err != nil {
			return err
		}

		//decrement following
		_, err = q.DecrementFollowing(c, arg.Username)
		if err != nil {
			return err
		}

		//increment follower
		_, err = q.DecrementFollower(c, arg.FollowUser)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}