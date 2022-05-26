package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func CreateRelations(t *testing.T) Relations {
	followedUser := CreateRandomUser(t)
	followerUser := CreateRandomUser(t)

	arg := CreateRelationsParams{
		FollowedUsername: followedUser.Username,
		FollowerUsername: followerUser.Username,
	}

	relations, err := testQueries.CreateRelations(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, relations)

	require.Equal(t, followedUser.Username, relations.FollowedUsername)
	require.Equal(t, followerUser.Username, relations.FollowerUsername)

	require.NotZero(t, relations.ID)
	require.NotZero(t, relations.CreatedAt)

	return relations
}

func TestCreateRelationsOK(t *testing.T) {
	CreateRelations(t)
}

func TestDeleteRelationsOK(t *testing.T) {
	relations := CreateRelations(t)

	followerUsername := relations.FollowerUsername
	followedUsername := relations.FollowedUsername

	delArg := DeleteRelationParams{
		FollowerUsername: followerUsername,
		FollowedUsername: followedUsername,
	}

	err := testQueries.DeleteRelation(context.Background(), delArg)

	require.NoError(t, err)

	getArg := GetRelationsParams{
		FollowerUsername: followerUsername,
		FollowedUsername: followedUsername,
	}

	getRelation, err := testQueries.GetRelations(context.Background(), getArg)

	require.Error(t, err)
	require.Empty(t, getRelation)
}

func TestGetFollowers(t *testing.T) {
	followedAccount := CreateRandomUser(t)
	var listFollowerAccount []Users

	for i := 0; i < 5; i++ {
		followerAccount := CreateRandomUser(t)
		listFollowerAccount = append(listFollowerAccount, followerAccount)
		arg := CreateRelationsParams{
			FollowerUsername: followerAccount.Username,
			FollowedUsername: followedAccount.Username,
		}
		relation, err := testQueries.CreateRelations(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, relation)
	}

	getFollowersArg := GetFollowerParams{
		FollowedUsername: followedAccount.Username,
		Limit: 5,
		Offset: 0,
	}
	followerList, err := testQueries.GetFollower(context.Background(), getFollowersArg)

	require.NoError(t, err)
	require.NotEmpty(t, followerList)
	require.Equal(t, 5, len(followerList))
	
	i := 4
	for _, rel := range followerList {
		require.Equal(t, followedAccount.Username, rel.FollowedUsername)
		require.Equal(t, listFollowerAccount[i].Username, rel.FollowerUsername)
		i--
	}
}

func TestGetFollowing(t *testing.T) {
	followerAccount := CreateRandomUser(t)
	var listFollowedAccount []Users

	for i := 0; i < 5; i++ {
		followedAccount := CreateRandomUser(t)
		listFollowedAccount = append(listFollowedAccount, followedAccount)
		arg := CreateRelationsParams{
			FollowerUsername: followerAccount.Username,
			FollowedUsername: followedAccount.Username,
		}
		relation, err := testQueries.CreateRelations(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, relation)
	}

	getFollowingArg := GetFollowingParams{
		FollowerUsername: followerAccount.Username,
		Limit: 5,
		Offset: 0,
	}
	followingList, err := testQueries.GetFollowing(context.Background(), getFollowingArg)

	require.NoError(t, err)
	require.NotEmpty(t, followingList)
	require.Equal(t, 5, len(followingList))
	
	i := 4
	for _, rel := range followingList {
		require.Equal(t, followerAccount.Username, rel.FollowerUsername)
		require.Equal(t, listFollowedAccount[i].Username, rel.FollowedUsername)
		i--
	}
}

func TestGetRelations(t *testing.T) {
	relation := CreateRelations(t)

	arg := GetRelationsParams{
		FollowerUsername: relation.FollowerUsername,
		FollowedUsername: relation.FollowedUsername,
	}

	getRel, err := testQueries.GetRelations(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, getRel)

	require.Equal(t, relation.FollowerUsername, getRel.FollowerUsername)
	require.Equal(t, relation.FollowedUsername, getRel.FollowedUsername)
}