package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"simplebank/util"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: "secret",
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.NotZero(t, user.CreatedAt)
	require.True(t, user.PasswordChangedAt.IsZero())

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	createdUser := createRandomUser(t)
	dbRecord, err := testQueries.GetUser(context.Background(), createdUser.Username)

	require.NoError(t, err)
	require.NotEmpty(t, dbRecord)

	require.Equal(t, createdUser.Username, dbRecord.Username)
	require.Equal(t, createdUser.HashedPassword, dbRecord.HashedPassword)
	require.Equal(t, createdUser.FullName, dbRecord.FullName)
	require.Equal(t, createdUser.Email, dbRecord.Email)
	require.WithinDuration(t, createdUser.CreatedAt, dbRecord.CreatedAt, time.Second)
	require.WithinDuration(t, createdUser.PasswordChangedAt, dbRecord.PasswordChangedAt, time.Second)
}
