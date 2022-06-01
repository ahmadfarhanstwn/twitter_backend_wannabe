package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ahmadfarhanstwn/twitter_wannabe/token"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(tokenMaker token.Paseto) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeaderKey)

		if len(authHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, fmt.Errorf("auth header is not provided"))
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, fmt.Errorf("auth header is invalid"))
			return
		}

		authType := strings.ToLower(fields[0])
		if authType != authorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, fmt.Errorf("unauthorized auth type : %v",authType))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, err)
			return
		}

		c.Set(authorizationPayloadKey, payload)
	}
}