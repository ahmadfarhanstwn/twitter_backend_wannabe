package database

import (
	"context"
	"testing"
	"time"

	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
	"github.com/stretchr/testify/require"
)

const tweets string = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

func CreateTweet(t *testing.T) Tweets {
	user := CreateRandomUser(t)

	arg := CreateTweetParams{
		Tweet: tweets,
		Username: user.Username,
	}

	tweet, err := testQueries.CreateTweet(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, tweet)
	
	require.Equal(t, tweets, tweet.Tweet)
	require.Equal(t, user.Username, tweet.Username)

	require.NotZero(t, tweet.ID)
	require.NotZero(t, tweet.CreatedAt)

	return tweet
}

func TestCreateTweet(t *testing.T) {
	CreateTweet(t)
}

func TestDeleteTweet(t *testing.T) {
	tweet := CreateTweet(t)

	idTweet := tweet.ID

	err := testQueries.DeleteTweet(context.Background(), idTweet)

	require.NoError(t, err)

	getTweet, err := testQueries.GetTweet(context.Background(), idTweet)

	require.Error(t, err)
	require.Empty(t, getTweet)
}

func TestGetListTweets(t *testing.T) {
	user := CreateRandomUser(t)

	var recentTweets []string

	for i := 0; i < 5; i++ {
		tweet := util.GetRandomString(20)
		recentTweets = append(recentTweets, tweet)
		arg := CreateTweetParams{
			Tweet: tweet,
			Username: user.Username,
		}
		tweeet, err := testQueries.CreateTweet(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, tweeet)
	}

	getTweetArg := GetListTweetsParams{
		Username: user.Username,
		Limit: 5,
		Offset: 0,
	}
	GetTweets, err := testQueries.GetListTweets(context.Background(), getTweetArg)
	require.NotEmpty(t, GetTweets)
	require.NoError(t, err)

	require.Equal(t, 5, len(GetTweets))

	i := 4
	for _, tw := range GetTweets{
		require.Equal(t, user.Username, tw.Username)
		require.Equal(t, recentTweets[i], tw.Tweet)
		i--
	}
}

func TestGetTweet(t *testing.T) {
	createdTweet := CreateTweet(t)

	getTweet, err := testQueries.GetTweet(context.Background(), createdTweet.ID)

	require.NoError(t, err)
	require.NotEmpty(t, getTweet)

	require.Equal(t, createdTweet.ID, getTweet.ID)
	require.Equal(t, createdTweet.Tweet, getTweet.Tweet)
	require.Equal(t, createdTweet.Username, getTweet.Username)
	require.WithinDuration(t, createdTweet.CreatedAt, getTweet.CreatedAt, time.Second)
}