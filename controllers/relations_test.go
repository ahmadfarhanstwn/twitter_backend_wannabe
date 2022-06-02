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
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestFollow(t *testing.T) {
	user, _ := randomUser(t)
	followUser, _ := randomUser(t)

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
				"follow_user" : followUser.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(followUser.Username)).Times(1).Return(followUser, nil)
				getRelationArg := database.GetRelationsParams{
					FollowerUsername: user.Username,
					FollowedUsername: followUser.Username,
				}
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Eq(getRelationArg)).Times(1).Return(database.Relations{}, sql.ErrNoRows)
				followInputArg := database.FollowInputArgs{
					Username: user.Username,
					FollowUser: followUser.Username,
				}
				transaction.EXPECT().FollowTx(gomock.Any(), gomock.Eq(followInputArg)).Times(1).Return(database.FollowInputResult{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"follow_usera" : followUser.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(0)
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Any()).Times(0)
				transaction.EXPECT().FollowTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "User Not Found",
			body: gin.H{
				"follow_user" : "followUser.Username",
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq("followUser.Username")).Times(1).Return(database.Users{}, sql.ErrNoRows)
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Any()).Times(0)
				transaction.EXPECT().FollowTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "User already followed",
			body: gin.H{
				"follow_user" : followUser.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(followUser.Username)).Times(1).Return(followUser, nil)
				getRelationArg := database.GetRelationsParams{
					FollowerUsername: user.Username,
					FollowedUsername: followUser.Username,
				}
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Eq(getRelationArg)).Times(1).Return(database.Relations{}, nil)
				transaction.EXPECT().FollowTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"follow_user" : followUser.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(followUser.Username)).Times(1).Return(followUser, nil)
				getRelationArg := database.GetRelationsParams{
					FollowerUsername: user.Username,
					FollowedUsername: followUser.Username,
				}
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Eq(getRelationArg)).Times(1).Return(database.Relations{}, sql.ErrNoRows)
				followInputArg := database.FollowInputArgs{
					Username: user.Username,
					FollowUser: followUser.Username,
				}
				transaction.EXPECT().FollowTx(gomock.Any(), gomock.Eq(followInputArg)).Times(1).Return(database.FollowInputResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases{
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

		url := "/follow"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestUnfollow(t *testing.T) {
	user, _ := randomUser(t)
	followUser, _ := randomUser(t)

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
				"follow_user" : followUser.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(followUser.Username)).Times(1).Return(followUser, nil)
				getRelationArg := database.GetRelationsParams{
					FollowerUsername: user.Username,
					FollowedUsername: followUser.Username,
				}
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Eq(getRelationArg)).Times(1).Return(database.Relations{}, nil)
				followInputArg := database.FollowInputArgs{
					Username: user.Username,
					FollowUser: followUser.Username,
				}
				transaction.EXPECT().UnfollowTx(gomock.Any(), gomock.Eq(followInputArg)).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"follow_usera" : followUser.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(0)
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Any()).Times(0)
				transaction.EXPECT().UnfollowTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "User Not Found",
			body: gin.H{
				"follow_user" : "followUser.Username",
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq("followUser.Username")).Times(1).Return(database.Users{}, sql.ErrNoRows)
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Any()).Times(0)
				transaction.EXPECT().UnfollowTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "User hasn't followed",
			body: gin.H{
				"follow_user" : followUser.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(followUser.Username)).Times(1).Return(followUser, nil)
				getRelationArg := database.GetRelationsParams{
					FollowerUsername: user.Username,
					FollowedUsername: followUser.Username,
				}
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Eq(getRelationArg)).Times(1).Return(database.Relations{}, sql.ErrNoRows)
				transaction.EXPECT().UnfollowTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"follow_user" : followUser.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(followUser.Username)).Times(1).Return(followUser, nil)
				getRelationArg := database.GetRelationsParams{
					FollowerUsername: user.Username,
					FollowedUsername: followUser.Username,
				}
				transaction.EXPECT().GetRelations(gomock.Any(), gomock.Eq(getRelationArg)).Times(1).Return(database.Relations{}, nil)
				followInputArg := database.FollowInputArgs{
					Username: user.Username,
					FollowUser: followUser.Username,
				}
				transaction.EXPECT().UnfollowTx(gomock.Any(), gomock.Eq(followInputArg)).Times(1).Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases{
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

		url := "/unfollow"
		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}