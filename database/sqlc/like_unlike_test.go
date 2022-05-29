package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLikeTx(t *testing.T) {
	dbt := NewTransaction(testDB)

	user := CreateRandomUser(t)
	tweet := CreateTweet(t)

	err := dbt.LikeTweetTx(context.Background(), CreateLikeRelationParams{
		Username: user.Username,
		TweetID: tweet.ID,
	})

	require.NoError(t, err)

	rel, err := dbt.GetLikeRelation(context.Background(), GetLikeRelationParams{
		Username: user.Username,
		TweetID: tweet.ID,
	})

	require.NoError(t, err)
	require.NotEmpty(t, rel)
	require.Equal(t, user.Username, rel.Username)
	require.Equal(t, tweet.ID, rel.TweetID)

	newTweet, err := dbt.GetTweet(context.Background(), tweet.ID)
	require.NoError(t, err)
	require.NotEmpty(t, newTweet)
	require.Equal(t, tweet.Likes.Int32+1, newTweet.Likes.Int32)
}

func TestUnlikeTx(t *testing.T) {
	dbt := NewTransaction(testDB)

	user := CreateRandomUser(t)
	tweet := CreateTweet(t)

	err := dbt.UnlikeTweetTx(context.Background(), DeleteLikeRelationParams{
		Username: user.Username,
		TweetID: tweet.ID,
	})
	require.NoError(t, err)

	_, err = dbt.GetLikeRelation(context.Background(), GetLikeRelationParams{
		Username: user.Username,
		TweetID: tweet.ID,
	})
	require.Error(t, err)

	newTweet, err := dbt.GetTweet(context.Background(), tweet.ID)
	require.NoError(t, err)
	require.NotEmpty(t, newTweet)
	require.Equal(t, tweet.Likes.Int32-1, newTweet.Likes.Int32)
}