package controllers

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ahmadfarhanstwn/twitter_wannabe/token"
	"github.com/stretchr/testify/require"
)

func AddAuth(
	t *testing.T,
	request *http.Request,
	paseto token.Paseto,
	authType string,
	username string,
	duration time.Duration,
) {
	token, err := paseto.CreateToken(username, duration)
	require.NoError(t, err)
	
	authHeader := fmt.Sprintf("%v %v", authType, token)
	request.Header.Set(authorizationHeaderKey, authHeader)
}