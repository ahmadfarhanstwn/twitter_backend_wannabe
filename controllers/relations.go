package controllers

import (
	"fmt"
	"net/http"

	database "github.com/ahmadfarhanstwn/twitter_wannabe/database/sqlc"
	"github.com/ahmadfarhanstwn/twitter_wannabe/token"
	"github.com/gin-gonic/gin"
)

type FollowReq struct {
	FollowUser string `json:"follow_user" binding:"required,min=1,max=30"`
}

func (s *Server) Follow(c *gin.Context) {
	var req FollowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	//check if want to follow user is exist
	_, err := s.transaction.GetUser(c, req.FollowUser)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrResponse(err.Error()))
		return
	}

	//check if already follow
	authHeader := c.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := database.GetRelationsParams{
		FollowerUsername: authHeader.Username,
		FollowedUsername: req.FollowUser,
	}
	_,err = s.transaction.GetRelations(c, arg)
	if err == nil {
		c.JSON(http.StatusCreated,gin.H{
			"error" : fmt.Sprintf("%v has already followed %v", authHeader.Username, req.FollowUser),
		})
		return
	}

	//////////////////////// FROM DBTRANSACTION ///////////////
	txArg := database.FollowInputArgs{
		Username: authHeader.Username,
		FollowUser: req.FollowUser,
	}

	_, err = s.transaction.FollowTx(c, txArg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"Message" : fmt.Sprintf("%v succesfully followed %v", authHeader.Username, req.FollowUser),
	})
}

func (s *Server) Unfollow(c *gin.Context) {
	var req FollowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	//check if want to unfollow user is exist
	_, err := s.transaction.GetUser(c, req.FollowUser)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrResponse(err.Error()))
		return
	}

	//check if its not following
	authHeader := c.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := database.GetRelationsParams{
		FollowerUsername: authHeader.Username,
		FollowedUsername: req.FollowUser,
	}
	_,err = s.transaction.GetRelations(c, arg)
	if err != nil {
		c.JSON(http.StatusCreated,gin.H{
			"error" : fmt.Sprintf("%v is not following %v", authHeader.Username, req.FollowUser),
		})
		return
	}

	txArg := database.FollowInputArgs{
		Username: authHeader.Username,
		FollowUser: req.FollowUser,
	}

	err = s.transaction.UnfollowTx(c, txArg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message" : fmt.Sprintf("%v succesfully unfollowed %v", authHeader.Username, req.FollowUser),
	})
}