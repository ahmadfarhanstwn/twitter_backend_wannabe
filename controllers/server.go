package controllers

import (
	"log"

	database "github.com/ahmadfarhanstwn/twitter_wannabe/database/sqlc"
	"github.com/ahmadfarhanstwn/twitter_wannabe/token"
	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "payload"
)

type Server struct {
	router *gin.Engine
	config util.Config
	transaction database.Transaction
	paseto token.Paseto
}

func NewServer(config util.Config, dbtx database.Transaction) (*Server, error) {
	paseto, err := token.NewPaseto(config.Access_Token)
	if err != nil {
		log.Fatal(err)
	}
	server := &Server{config: config, transaction: dbtx, paseto: *paseto}
	server.SetupRouter()
	return server, nil
}

func (s *Server) SetupRouter(){
	router := gin.Default()

	// out of auth middleware
	router.POST("/register", s.SignUp)
	router.POST("/login", s.Login)

	authRouter := router.Group("/").Use(AuthMiddleware(s.paseto))

	//user
	authRouter.GET("/profile", s.GetUserProfile)
	authRouter.PUT("/password", s.UpdatePassword)
	authRouter.PUT("/email", s.UpdateEmail)
	authRouter.PUT("/name", s.UpdateName)
	authRouter.GET("/followers", s.GetFollowersList)
	authRouter.GET("/following", s.GetFollowingList)

	//tweets
	authRouter.POST("/tweet", s.CreateTweet)
	authRouter.DELETE("/tweet", s.DeleteTweet)
	authRouter.GET("/tweet", s.GetTweet)
	authRouter.POST("/like", s.LikeTweet)
	authRouter.DELETE("/unlike", s.UnlikeTweet)
	authRouter.GET("/feeds", s.GetFeeds)

	//relations
	authRouter.POST("/follow", s.Follow)
	authRouter.DELETE("/unfollow", s.Unfollow)

	s.router = router
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func ErrResponse(error string) gin.H {
	return gin.H{"error": error}
}