package database

import (
	"context"
	"testing"
	"time"

	"github.com/ahmadfarhanstwn/twitter_wannabe/util"
	"github.com/stretchr/testify/require"
)

func CreateRandomUser(t *testing.T) Users {
	arg := CreateUserParams{
		Username: util.GetRandomString(8),
		HashedPassword: util.GetRandomString(6),
		Name: util.GetRandomString(8),
		Email: util.GetRandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Name, user.Name)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, int32(0), user.FollowersCount.Int32)
	require.Equal(t, int32(0), user.FollowingCount.Int32)

	require.NotZero(t, user.CreatedAt)

	require.Zero(t, user.ChangedPasswordAt)

	return user
}

func TestCreateUserOK(t *testing.T) {
	CreateRandomUser(t)
}

func TestGetUserOK(t *testing.T) {
	createdUser := CreateRandomUser(t)
	fetchedUser, err := testQueries.GetUser(context.Background(), createdUser.Username)

	require.NoError(t, err)
	require.NotEmpty(t, fetchedUser)

	require.Equal(t, createdUser.Username, fetchedUser.Username)
	require.Equal(t, createdUser.Email, fetchedUser.Email)
	require.Equal(t, createdUser.Name, fetchedUser.Name)
	require.Equal(t, createdUser.HashedPassword, fetchedUser.HashedPassword)
	require.Equal(t, createdUser.FollowersCount, fetchedUser.FollowersCount)
	require.Equal(t, createdUser.FollowingCount, fetchedUser.FollowingCount)

	require.WithinDuration(t, createdUser.ChangedPasswordAt, fetchedUser.ChangedPasswordAt, time.Second)
	require.WithinDuration(t, createdUser.CreatedAt, fetchedUser.CreatedAt, time.Second)
}

func TestUpdateUserOK(t *testing.T) {
	user := CreateRandomUser(t)
	newEmail := util.GetRandomEmail()
	newPassword := util.GetRandomString(6)
	newName := util.GetRandomString(8)

	params := UpdateUserParams{
		Email: newEmail,
		HashedPassword: newPassword,
		Name: newName,
		Username: user.Username,
	}

	updatedUser, err := testQueries.UpdateUser(context.Background(), params)

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, newPassword, updatedUser.HashedPassword)
	require.Equal(t, newName, updatedUser.Name)
}

func TestIncrementFollowing(t *testing.T) {
	user := CreateRandomUser(t)

	updatedUser, err := testQueries.IncrementFollowing(context.Background(), user.Username)
	
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, updatedUser.FollowingCount.Int32, user.FollowingCount.Int32+int32(1))
}

func TestDecrementFollowing(t *testing.T) {
	user := CreateRandomUser(t)

	updatedUser, err := testQueries.DecrementFollowing(context.Background(), user.Username)
	
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, updatedUser.FollowingCount.Int32, user.FollowingCount.Int32-int32(1))
}

func TestIncrementFollower(t *testing.T) {
	user := CreateRandomUser(t)

	updatedUser, err := testQueries.IncrementFollower(context.Background(), user.Username)
	
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, updatedUser.FollowersCount.Int32, user.FollowersCount.Int32+int32(1))
}

func TestDecrementFollower(t *testing.T) {
	user := CreateRandomUser(t)

	updatedUser, err := testQueries.DecrementFollower(context.Background(), user.Username)
	
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, updatedUser.FollowersCount.Int32, user.FollowersCount.Int32-int32(1))
}