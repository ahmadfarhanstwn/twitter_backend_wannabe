package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dbmock "github.com/ahmadfarhanstwn/twitter_wannabe/database/mock"
	database "github.com/ahmadfarhanstwn/twitter_wannabe/database/sqlc"
	"github.com/ahmadfarhanstwn/twitter_wannabe/token"
	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func randomTweets(user database.Users) database.Tweets {
	return database.Tweets{
		ID: 1,
		Tweet: util.GetRandomString(15),
		Username: user.Username,
		Likes: sql.NullInt32{Int32: 0, Valid: true},
		CreatedAt: time.Now(),
	}
}

func TestCreateTweet(t *testing.T) {
	user, _ := randomUser(t)
	tweet := randomTweets(user)

	testcases := []struct{
		name string
		body gin.H
		setupAuth func(t *testing.T, request *http.Request, paseto token.Paseto)
		buildStubs func(transaction *dbmock.MockTransaction)
		checkResponse func(t *testing.T,recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"tweet" : tweet.Tweet,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.CreateTweetParams{
					Tweet: tweet.Tweet,
					Username: user.Username,
				}
				transaction.EXPECT().CreateTweet(gomock.Any(), gomock.Eq(arg)).Times(1).Return(tweet, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"dongeng" : tweet.Tweet,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().CreateTweet(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"tweet" : tweet.Tweet,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.CreateTweetParams{
					Tweet: tweet.Tweet,
					Username: user.Username,
				}
				transaction.EXPECT().CreateTweet(gomock.Any(), gomock.Eq(arg)).Times(1).Return(database.Tweets{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases {
		//create controller
		controller := gomock.NewController(t)
		defer controller.Finish()

		//create mock transaction
		transaction := dbmock.NewMockTransaction(controller)
		testcase.buildStubs(transaction)

		// create test server
		server := NewTestServer(t, transaction)
		recorder := httptest.NewRecorder()

		// marshal/read body params
		data, err := json.Marshal(testcase.body)
		require.NoError(t, err)

		url := "/tweet"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestDeleteTweet(t *testing.T) {
	user, _ := randomUser(t)
	tweet := randomTweets(user)

	testcases := []struct{
		name string
		body gin.H
		setupAuth func(t *testing.T, request *http.Request, paseto token.Paseto)
		buildStubs func(transaction *dbmock.MockTransaction)
		checkResponse func(t *testing.T,recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(tweet, nil)
				transaction.EXPECT().DeleteTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"ide" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(0)
				transaction.EXPECT().DeleteTweet(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "User not found",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(database.Tweets{}, sql.ErrNoRows)
				transaction.EXPECT().DeleteTweet(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Internal server error(get)",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(database.Tweets{}, sql.ErrConnDone)
				transaction.EXPECT().DeleteTweet(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Internal server error (delete)",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(tweet, nil)
				transaction.EXPECT().DeleteTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases {
		//create controller
		controller := gomock.NewController(t)
		defer controller.Finish()

		//create mock transaction
		transaction := dbmock.NewMockTransaction(controller)
		testcase.buildStubs(transaction)

		// create test server
		server := NewTestServer(t, transaction)
		recorder := httptest.NewRecorder()

		// marshal/read body params
		data, err := json.Marshal(testcase.body)
		require.NoError(t, err)

		url := "/tweet"
		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestGetTweet(t *testing.T) {
	user, _ := randomUser(t)
	tweet := randomTweets(user)

	testcases := []struct{
		name string
		body gin.H
		setupAuth func(t *testing.T, request *http.Request, paseto token.Paseto)
		buildStubs func(transaction *dbmock.MockTransaction)
		checkResponse func(t *testing.T,recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(tweet, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"ide" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "User not found",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(database.Tweets{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetTweet(gomock.Any(), gomock.Eq(tweet.ID)).Times(1).Return(database.Tweets{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases {
		//create controller
		controller := gomock.NewController(t)
		defer controller.Finish()

		//create mock transaction
		transaction := dbmock.NewMockTransaction(controller)
		testcase.buildStubs(transaction)

		// create test server
		server := NewTestServer(t, transaction)
		recorder := httptest.NewRecorder()

		// marshal/read body params
		data, err := json.Marshal(testcase.body)
		require.NoError(t, err)

		url := "/tweet"
		req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestLikeTweet(t *testing.T) {
	user, _ := randomUser(t)
	tweet := randomTweets(user)

	testcases := []struct{
		name string
		body gin.H
		setupAuth func(t *testing.T, request *http.Request, paseto token.Paseto)
		buildStubs func(transaction *dbmock.MockTransaction)
		checkResponse func(t *testing.T,recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				getArg := database.GetLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				createArg := database.CreateLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				transaction.EXPECT().GetLikeRelation(gomock.Any(), gomock.Eq(getArg)).Times(1).Return(database.LikeRelations{}, sql.ErrNoRows)
				transaction.EXPECT().LikeTweetTx(gomock.Any(), gomock.Eq(createArg)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"ide" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				getArg := database.GetLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				createArg := database.CreateLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				transaction.EXPECT().GetLikeRelation(gomock.Any(), gomock.Eq(getArg)).Times(0)
				transaction.EXPECT().LikeTweetTx(gomock.Any(), gomock.Eq(createArg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "User already liked the post",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				getArg := database.GetLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				createArg := database.CreateLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				transaction.EXPECT().GetLikeRelation(gomock.Any(), gomock.Eq(getArg)).Times(1).Return(database.LikeRelations{}, nil)
				transaction.EXPECT().LikeTweetTx(gomock.Any(), gomock.Eq(createArg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				getArg := database.GetLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				createArg := database.CreateLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				transaction.EXPECT().GetLikeRelation(gomock.Any(), gomock.Eq(getArg)).Times(1).Return(database.LikeRelations{}, sql.ErrNoRows)
				transaction.EXPECT().LikeTweetTx(gomock.Any(), gomock.Eq(createArg)).Times(1).Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases {
		//create controller
		controller := gomock.NewController(t)
		defer controller.Finish()

		//create mock transaction
		transaction := dbmock.NewMockTransaction(controller)
		testcase.buildStubs(transaction)

		// create test server
		server := NewTestServer(t, transaction)
		recorder := httptest.NewRecorder()

		// marshal/read body params
		data, err := json.Marshal(testcase.body)
		require.NoError(t, err)

		url := "/like"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestUnlikeTweet(t *testing.T) {
	user, _ := randomUser(t)
	tweet := randomTweets(user)

	testcases := []struct{
		name string
		body gin.H
		setupAuth func(t *testing.T, request *http.Request, paseto token.Paseto)
		buildStubs func(transaction *dbmock.MockTransaction)
		checkResponse func(t *testing.T,recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				getArg := database.GetLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				deleteArg := database.DeleteLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				transaction.EXPECT().GetLikeRelation(gomock.Any(), gomock.Eq(getArg)).Times(1).Return(database.LikeRelations{}, nil)
				transaction.EXPECT().UnlikeTweetTx(gomock.Any(), gomock.Eq(deleteArg)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"ide" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetLikeRelation(gomock.Any(), gomock.Any()).Times(0)
				transaction.EXPECT().UnlikeTweetTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "User hasn't liked the tweet",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				getArg := database.GetLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				transaction.EXPECT().GetLikeRelation(gomock.Any(), gomock.Eq(getArg)).Times(1).Return(database.LikeRelations{}, sql.ErrNoRows)
				transaction.EXPECT().UnlikeTweetTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				getArg := database.GetLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				deleteArg := database.DeleteLikeRelationParams{
					Username: user.Username,
					TweetID: tweet.ID,
				}
				transaction.EXPECT().GetLikeRelation(gomock.Any(), gomock.Eq(getArg)).Times(1).Return(database.LikeRelations{}, nil)
				transaction.EXPECT().UnlikeTweetTx(gomock.Any(), gomock.Eq(deleteArg)).Times(1).Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases {
		//create controller
		controller := gomock.NewController(t)
		defer controller.Finish()

		//create mock transaction
		transaction := dbmock.NewMockTransaction(controller)
		testcase.buildStubs(transaction)

		// create test server
		server := NewTestServer(t, transaction)
		recorder := httptest.NewRecorder()

		// marshal/read body params
		data, err := json.Marshal(testcase.body)
		require.NoError(t, err)

		url := "/unlike"
		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}