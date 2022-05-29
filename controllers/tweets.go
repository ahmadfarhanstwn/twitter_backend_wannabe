package controllers

import (
	"fmt"
	"net/http"

	database "github.com/ahmadfarhanstwn/twitter_wannabe/database/sqlc"
	"github.com/ahmadfarhanstwn/twitter_wannabe/token"
	"github.com/gin-gonic/gin"
)

type CreateTweetRequest struct {
	Tweet string `json:"tweet" binding:"required"`
}

func (s *Server) CreateTweet(c *gin.Context) {
	var req CreateTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	authHeader := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := database.CreateTweetParams{
		Username: authHeader.Username,
		Tweet: req.Tweet,
	}
	createdTweet, err := s.transaction.CreateTweet(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, createdTweet)
}

type DeleteGetAndLikeTweetRequest struct {
	ID int64 `json:"id" binding:"required"`
}

func (s *Server) DeleteTweet(c *gin.Context) {
	var req DeleteGetAndLikeTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	err := s.transaction.DeleteTweet(c, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message" : fmt.Sprintf("Tweet with ID %v has succesfully been deleted", req.ID),
	})
}

func (s *Server) GetTweet(c *gin.Context) {
	var req DeleteGetAndLikeTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	tweet, err := s.transaction.GetTweet(c, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, tweet)
}

//TODO : SHOULD IMPLEMENT TRANSACTION ISOLATIONS
func (s *Server) LikeTweet(c *gin.Context) {
	var req DeleteGetAndLikeTweetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	authHeader := c.MustGet(authorizationPayloadKey).(*token.Payload)

	//make sure user hasn't liked the tweet
	_, err := s.transaction.GetLikeRelation(c, database.GetLikeRelationParams{
		Username: authHeader.Username,
		TweetID: req.ID,
	})
	if err == nil {
		c.JSON(http.StatusCreated, gin.H{
			"error" : fmt.Sprintf("%v has already liked tweet %v", authHeader.Username, req.ID),
		})
	}

	//TRANSACTION
	txArg := database.CreateLikeRelationParams{
		Username: authHeader.Username,
		TweetID: req.ID,
	}
	err = s.transaction.LikeTweetTx(c, txArg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%v liked tweet %v", authHeader.Username, req.ID),
	})
}

func (s *Server) UnlikeTweet(c *gin.Context) {
	var req DeleteGetAndLikeTweetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	authHeader := c.MustGet(authorizationPayloadKey).(*token.Payload)

	//make sure user has liked the tweet
	_, err := s.transaction.GetLikeRelation(c, database.GetLikeRelationParams{
		Username: authHeader.Username,
		TweetID: req.ID,
	})
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{
			"error" : fmt.Sprintf("%v hasn't liked tweet %v", authHeader.Username, req.ID),
		})
	}

	txArg := database.DeleteLikeRelationParams{
		Username: authHeader.Username,
		TweetID: req.ID,
	}
	err = s.transaction.UnlikeTweetTx(c, txArg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%v unliked tweet %v", authHeader.Username, req.ID),
	})
}