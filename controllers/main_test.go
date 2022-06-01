package controllers

import (
	"os"
	"testing"
	"time"

	database "github.com/ahmadfarhanstwn/twitter_wannabe/database/sqlc"
	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func NewTestServer(t *testing.T, db database.Transaction) *Server {
	config := util.Config{
		Access_Token: util.GetRandomString(32),
		Token_Duration: time.Minute,
	}

	server, err := NewServer(config, db)
	require.NoError(t, err)
	
	return server
}