package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateLikeRelations(t *testing.T) {
	tweet := CreateTweet(t)

	arg := CreateLikeRelationParams{
		Username: tweet.Username,
		TweetID: tweet.ID,
	}

	likeRelation, err := testQueries.CreateLikeRelation(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, likeRelation)

	require.Equal(t, tweet.Username, likeRelation.Username)
	require.Equal(t, tweet.ID, likeRelation.TweetID)
	require.NotZero(t, likeRelation.ID)
	require.NotZero(t, likeRelation.CreatedAt)
}