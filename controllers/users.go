package controllers

import (
	"database/sql"
	"net/http"

	database "github.com/ahmadfarhanstwn/twitter_wannabe/database/sqlc"
	"github.com/ahmadfarhanstwn/twitter_wannabe/token"
	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type signUpForm struct {
	Username string `json:"username" binding:"required,min=1,max=15"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=30"`
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type signUpAndUpdateResp struct {
	Username string `json:"username"`
	Email string `json:"email"`
	Name string `json:"name"`
	Followers_Count int32 `json:"followers_count"`
	Following_Count int32 `json:"following_count"`
}

func (s *Server) SignUp(c *gin.Context) {
	var req signUpForm
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	arg := database.CreateUserParams{
		Username: req.Username,
		Email: req.Email,
		HashedPassword: hashedPassword,
		Name: req.Name,
	}

	user, err := s.transaction.CreateUser(c, arg)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				c.JSON(http.StatusForbidden, ErrResponse(err.Error()))
				return
			}
		}
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	resp := signUpAndUpdateResp{
		Username: user.Username,
		Email: user.Email,
		Name: user.Name,
		Followers_Count: user.FollowersCount.Int32,
		Following_Count: user.FollowingCount.Int32,
	}

	c.JSON(http.StatusOK, resp)
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=1,max=15"`
	Password string `json:"password" binding:"required,min=8,max=30"`
}

type LoginResp struct {
	Username string `json:"username"`
	CreatedToken string `json:"token"`
}

func (s *Server) Login(c *gin.Context) {
	var loginReq LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	user, err := s.transaction.GetUser(c, loginReq.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, ErrResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	err = util.CheckHashPassword(loginReq.Password, user.HashedPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrResponse(err.Error()))
		return
	}

	token, err := s.paseto.CreateToken(user.Username, s.config.Token_Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	resp := LoginResp{
		Username: loginReq.Username,
		CreatedToken: token,
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) GetUserProfile(c *gin.Context) {
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)
	user, err := s.transaction.GetUser(c, authPayload.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, user)
}

type UpdatePasswordReq struct {
	NewPassword string `json:"new_password" binding:"required,min=8,max=30"`
}

func (s *Server) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	hashedPassword, err := util.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := database.UpdatePasswordParams{
		Username: authPayload.Username,
		HashedPassword: hashedPassword,
	}
	user, err := s.transaction.UpdatePassword(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	resp := signUpAndUpdateResp{
		Username: user.Username,
		Email: user.Email,
		Name: user.Name,
		Followers_Count: user.FollowersCount.Int32,
		Following_Count: user.FollowingCount.Int32,
	}

	c.JSON(http.StatusOK, resp)
}

type UpdateEmailReq struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

func (s *Server) UpdateEmail(c *gin.Context) {
	var req UpdateEmailReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := database.UpdateEmailParams{
		Username: authPayload.Username,
		Email: req.NewEmail,
	}
	user, err := s.transaction.UpdateEmail(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	resp := signUpAndUpdateResp{
		Username: user.Username,
		Email: user.Email,
		Name: user.Name,
		Followers_Count: user.FollowersCount.Int32,
		Following_Count: user.FollowingCount.Int32,
	}

	c.JSON(http.StatusOK, resp)
}

type UpdateNameReq struct {
	NewName string `json:"new_name" binding:"required,min=1,max=50"`
}

func (s *Server) UpdateName(c *gin.Context) {
	var req UpdateNameReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := database.UpdateNameParams{
		Username: authPayload.Username,
		Name: req.NewName,
	}
	user, err := s.transaction.UpdateName(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	resp := signUpAndUpdateResp{
		Username: user.Username,
		Email: user.Email,
		Name: user.Name,
		Followers_Count: user.FollowersCount.Int32,
		Following_Count: user.FollowingCount.Int32,
	}

	c.JSON(http.StatusOK, resp)
}

type GetListRequest struct {
	PageSize int32 `json:"page_size" binding:"required,min=5"`
	PageId int32 `json:"page_id" binding:"required,min=1"`
}

func (s *Server) GetFollowersList(c *gin.Context) {
	var req GetListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	var resp []signUpAndUpdateResp

	authHeader := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := database.GetFollowerParams{
		FollowedUsername: authHeader.Username,
		Limit: req.PageSize,
		Offset: req.PageId,
	}

	followers, err := s.transaction.GetFollower(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	for _,f := range followers {
		follower, err := s.transaction.GetUser(c, f.FollowerUsername)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrResponse(err.Error()))
			return
		}
		r := signUpAndUpdateResp{
			Username: follower.Username,
			Email: follower.Email,
			Name: follower.Name,
			Followers_Count: follower.FollowersCount.Int32,
			Following_Count: follower.FollowingCount.Int32,
		}

		resp = append(resp, r)
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) GetFollowingList(c *gin.Context) {
	var req GetListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrResponse(err.Error()))
		return
	}

	var resp []signUpAndUpdateResp

	authHeader := c.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := database.GetFollowingParams{
		FollowerUsername: authHeader.Username,
		Limit: req.PageSize,
		Offset: req.PageId,
	}

	followings, err := s.transaction.GetFollowing(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrResponse(err.Error()))
		return
	}

	for _,f := range followings {
		following, err := s.transaction.GetUser(c, f.FollowedUsername)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrResponse(err.Error()))
			return
		}
		r := signUpAndUpdateResp{
			Username: following.Username,
			Email: following.Email,
			Name: following.Name,
			Followers_Count: following.FollowersCount.Int32,
			Following_Count: following.FollowingCount.Int32,
		}

		resp = append(resp, r)
	}

	c.JSON(http.StatusOK, resp)
}