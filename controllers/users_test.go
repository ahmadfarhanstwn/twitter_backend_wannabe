package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ahmadfarhanstwn/twitter_wannabe/database/mock"
	database "github.com/ahmadfarhanstwn/twitter_wannabe/database/sqlc"
	"github.com/ahmadfarhanstwn/twitter_wannabe/token"
	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func randomUser(t *testing.T) (database.Users, string) {
	password := util.GetRandomString(8)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	return database.Users{
		Username: util.GetRandomString(8),
		Email: util.GetRandomEmail(),
		HashedPassword: hashedPassword,
		Name: util.GetRandomString(8),
	}, password
}

type eqCreateUserParamsMatcher struct{
	arg database.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(database.CreateUserParams)
	if !ok {
		return false
	}
	err := util.CheckHashPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string{
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func eqCreateUserParams(arg database.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
} 

func TestSignUp(t *testing.T) {
	user, password := randomUser(t)

	testcases := []struct{
		name string
		bodyParams gin.H
		buildStubs func(transaction *dbmock.MockTransaction)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			bodyParams: gin.H{
				"username" : user.Username,
				"email" : user.Email,
				"password" : password,
				"name" : user.Name,
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.CreateUserParams{
					Username: user.Username,
					Email: user.Email,
					Name: user.Name,
				}
				transaction.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(arg, password)).Times(1).Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			bodyParams: gin.H{
				"usernamex" : user.Username,
				"email" : user.Email,
				"password" : password,
				"name" : user.Name,
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.CreateUserParams{
					Username: user.Username,
					Email: user.Email,
					Name: user.Name,
				}
				transaction.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(arg, password)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal Server Error",
			bodyParams: gin.H{
				"username" : user.Username,
				"email" : user.Email,
				"password" : password,
				"name" : user.Name,
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(1).Return(database.Users{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "duplicate username",
			bodyParams: gin.H{
				"username": user.Username,
				"email": user.Email,
				"name": user.Name,
				"password": password,
			},
			buildStubs: func(store *dbmock.MockTransaction) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(1).Return(database.Users{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "email invalid",
			bodyParams: gin.H{
				"username": user.Username,
				"email": "user.Email",
				"name": user.Name,
				"password": password,
			},
			buildStubs: func(store *dbmock.MockTransaction) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "password too short",
			bodyParams: gin.H{
				"username": user.Username,
				"email": user.Email,
				"name": user.Name,
				"password": "joko",
			},
			buildStubs: func(store *dbmock.MockTransaction) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
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
			data, err := json.Marshal(testcase.bodyParams)
			require.NoError(t, err)

			url := "/register"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			testcase.checkResponse(t, recorder)
		})
	}
}

func TestLogin(t *testing.T) {
	user, password := randomUser(t)

	testcases := []struct{
		name string
		body gin.H
		buildStubs func(transaction *dbmock.MockTransaction)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username" : user.Username,
				"password" : password,
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"usernamex" : user.Username,
				"password" : password,
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "User not found",
			body: gin.H{
				"username" : "user.Usernamex",
				"password" : password,
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(1).Return(database.Users{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "Internal Server Error",
			body: gin.H{
				"username" : user.Username,
				"password" : password,
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(1).Return(database.Users{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Invalid Password",
			body: gin.H{
				"username" : user.Username,
				"password" : "password",
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(1).Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Invalid Username",
			body: gin.H{
				"username" : "jokowidodoadalahpresidenterbaik",
				"password" : password,
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
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

			url := "/login"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			testcase.checkResponse(recorder)
		})
	}
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, account database.Users) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotUser database.Users
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, account, gotUser)
}

func TestGetProfile(t *testing.T) {
	user, _ := randomUser(t)

	testcases := []struct{
		name string
		setupAuth func(t *testing.T, request *http.Request, paseto token.Paseto)
		buildStubs func(transaction *dbmock.MockTransaction)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(1).Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "Internal Server Error",
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(1).Return(database.Users{}, sql.ErrConnDone)
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

		url := "/profile"
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestUpdateName(t *testing.T) {
	user, _:= randomUser(t)
	NewName := util.GetRandomString(8)

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
				"new_name" : NewName,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.UpdateNameParams{
					Username: user.Username,
					Name: NewName,
				}
				transaction.EXPECT().UpdateName(gomock.Any(), gomock.Eq(arg)).Times(1).Return(database.Users{
					Username: user.Username,
					Email: user.Email,
					HashedPassword: user.HashedPassword,
					Name: NewName,
					FollowersCount: user.FollowersCount,
					FollowingCount: user.FollowingCount,
					ChangedPasswordAt: user.ChangedPasswordAt,
					CreatedAt: user.CreatedAt,
				}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"new_emaaill" : NewName,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().UpdateName(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"new_name" : NewName,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.UpdateNameParams{
					Username: user.Username,
					Name: NewName,
				}
				transaction.EXPECT().UpdateName(gomock.Any(), gomock.Eq(arg)).Times(1).Return(database.Users{}, sql.ErrConnDone)
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

		url := "/name"
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestUpdateEmail(t *testing.T) {
	user, _:= randomUser(t)
	newEmail := util.GetRandomEmail()

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
				"new_email" : newEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.UpdateEmailParams{
					Username: user.Username,
					Email: newEmail,
				}
				transaction.EXPECT().UpdateEmail(gomock.Any(), gomock.Eq(arg)).Times(1).Return(database.Users{
					Username: user.Username,
					Email: newEmail,
					HashedPassword: user.HashedPassword,
					Name: user.Name,
					FollowersCount: user.FollowersCount,
					FollowingCount: user.FollowingCount,
					ChangedPasswordAt: user.ChangedPasswordAt,
					CreatedAt: user.CreatedAt,
				}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"new_emaaill" : newEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().UpdateEmail(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"new_email" : newEmail,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.UpdateEmailParams{
					Username: user.Username,
					Email: newEmail,
				}
				transaction.EXPECT().UpdateEmail(gomock.Any(), gomock.Eq(arg)).Times(1).Return(database.Users{}, sql.ErrConnDone)
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

		url := "/email"
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestGetFollowers(t *testing.T) {
	user, _ := randomUser(t)

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
				"page_size" : 5,
				"page_id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.GetFollowerParams{
					FollowedUsername: user.Username,
					Limit: 5,
					Offset: 1,
				}
				transaction.EXPECT().GetFollower(gomock.Any(), gomock.Eq(arg)).Times(1).Return([]database.Relations{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"page_sizee" : 5,
				"page_id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetFollower(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"page_size" : 5,
				"page_id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetFollower(gomock.Any(), gomock.Any()).Times(1).Return([]database.Relations{}, sql.ErrConnDone)
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

		url := "/followers"
		req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}

func TestGetFollowing(t *testing.T) {
	user, _ := randomUser(t)

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
				"page_size" : 5,
				"page_id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				arg := database.GetFollowingParams{
					FollowerUsername: user.Username,
					Limit: 5,
					Offset: 1,
				}
				transaction.EXPECT().GetFollowing(gomock.Any(), gomock.Eq(arg)).Times(1).Return([]database.Relations{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"page_sizee" : 5,
				"page_id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetFollowing(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal server error",
			body: gin.H{
				"page_size" : 5,
				"page_id" : 1,
			},
			setupAuth: func(t *testing.T, request *http.Request, paseto token.Paseto) {
				AddAuth(t, request, paseto, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(transaction *dbmock.MockTransaction) {
				transaction.EXPECT().GetFollowing(gomock.Any(), gomock.Any()).Times(1).Return([]database.Relations{}, sql.ErrConnDone)
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

		url := "/following"
		req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
		require.NoError(t, err)

		testcase.setupAuth(t, req, server.paseto)
		server.router.ServeHTTP(recorder, req)
		testcase.checkResponse(t,recorder)
	}
}