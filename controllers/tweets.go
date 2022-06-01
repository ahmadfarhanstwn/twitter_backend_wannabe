package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"

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

	//check if tweet exist
	_, err := s.transaction.GetTweet(c, req.ID)
	if err != nil {
		if err == sql.ErrNoRows{
			c.JSON(http.StatusNotFound, ErrResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	err = s.transaction.DeleteTweet(c, req.ID)
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
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, ErrResponse(err.Error()))
			return
		}
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
	if err != sql.ErrNoRows {
		c.JSON(http.StatusCreated, gin.H{
			"error" : fmt.Sprintf("%v has already liked tweet %v", authHeader.Username, req.ID),
		})
		return
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
		c.JSON(http.StatusNotFound, gin.H{
			"error" : fmt.Sprintf("%v hasn't liked tweet %v", authHeader.Username, req.ID),
		})
		return
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

func (s *Server) GetFeeds(c *gin.Context) {
	authHeader := c.MustGet(authorizationPayloadKey).(*token.Payload)

	var feeds []database.Tweets

	// get following list
	relations, err := s.transaction.GetFollowing(c, database.GetFollowingParams{
		FollowerUsername: authHeader.Username,
		Limit: 10000,
		Offset: 0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	for _, relation := range relations {
		// get tweets list from each following user
		tweets, err := s.transaction.GetListTweets(c,database.GetListTweetsParams{
			Username: relation.FollowedUsername,
			Limit: 100,
			Offset: 0,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
			return
		}
		feeds = append(feeds, tweets...)
	}

	// sort feeds by recent tweets
	sort.Slice(feeds, func(i, j int) bool {
		return feeds[i].ID > feeds[j].ID
	})

	c.JSON(http.StatusOK, feeds)
}