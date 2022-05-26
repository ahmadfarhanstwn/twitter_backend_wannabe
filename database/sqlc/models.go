// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0

package database

import (
	"database/sql"
	"time"
)

type Relations struct {
	ID               int64     `json:"id"`
	FollowerUsername string    `json:"follower_username"`
	FollowedUsername string    `json:"followed_username"`
	CreatedAt        time.Time `json:"created_at"`
}

type Tweets struct {
	ID        int64     `json:"id"`
	Tweet     string    `json:"tweet"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Users struct {
	Username          string        `json:"username"`
	Email             string        `json:"email"`
	HashedPassword    string        `json:"hashed_password"`
	Name              string        `json:"name"`
	FollowersCount    sql.NullInt32 `json:"followers_count"`
	FollowingCount    sql.NullInt32 `json:"following_count"`
	ChangedPasswordAt time.Time     `json:"changed_password_at"`
	CreatedAt         time.Time     `json:"created_at"`
}
