package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFollowTx(t *testing.T) {
	dbt := NewTransaction(testDB)

	followerUser := CreateRandomUser(t)
	followedUser := CreateRandomUser(t)

	for i := 0; i < 5; i++ {
		user1Before, _ := dbt.GetUser(context.Background(), followerUser.Username)
		user2Before, _ := dbt.GetUser(context.Background(), followedUser.Username)
		result, err := dbt.FollowTx(context.Background(), FollowInputArgs{
			Username: followerUser.Username,
			FollowUser: followedUser.Username,
		})
		require.NoError(t, err)
		require.NotEmpty(t, result)

		user1After, _ := dbt.GetUser(context.Background(), followerUser.Username)
		user2After, _ := dbt.GetUser(context.Background(), followedUser.Username)

		require.Equal(t, user1Before.FollowingCount.Int32+int32(1), user1After.FollowingCount.Int32)
		require.Equal(t, user2Before.FollowersCount.Int32+int32(1), user2After.FollowersCount.Int32)
		require.Equal(t, followerUser.Username, result.FollowerUser)
		require.Equal(t, followedUser.Username, result.FollowedUser)

		arg := GetRelationsParams{
			FollowerUsername: followerUser.Username,
			FollowedUsername: followedUser.Username,
		}
		_, err = dbt.GetRelations(context.Background(), arg)
		require.NoError(t, err)
	}
}

func TestUnfollowTx(t *testing.T) {
	dbt := NewTransaction(testDB)

	followerUser := CreateRandomUser(t)
	followedUser := CreateRandomUser(t)

	for i := 0; i < 5; i++ {
		followResult, err := dbt.FollowTx(context.Background(), FollowInputArgs{
			Username: followerUser.Username,
			FollowUser: followedUser.Username,
		})
		// err := <- errs
		require.NoError(t, err)

		// result := <- results
		require.NotEmpty(t, followResult)

		user1Before, _ := dbt.GetUser(context.Background(), followerUser.Username)
		user2Before, _ := dbt.GetUser(context.Background(), followedUser.Username)

		err = dbt.UnfollowTx(context.Background(), FollowInputArgs{
			Username: followerUser.Username,
			FollowUser: followedUser.Username,
		})
		user1After, _ := dbt.GetUser(context.Background(), followerUser.Username)
		user2After, _ := dbt.GetUser(context.Background(), followedUser.Username)

		require.Equal(t, user1Before.FollowingCount.Int32-int32(1), user1After.FollowingCount.Int32)
		require.Equal(t, user2Before.FollowersCount.Int32-int32(1), user2After.FollowersCount.Int32)
		require.NoError(t, err)
		// require.NotEmpty(t, unfollowResult)

		arg := GetRelationsParams{
			FollowerUsername: followerUser.Username,
			FollowedUsername: followedUser.Username,
		}
		_, err = dbt.GetRelations(context.Background(), arg)
		require.Error(t, err)
	}
}