package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"simplebank/util"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testStore.CreateUser(context.Background(), arg)
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
	dbRecord, err := testStore.GetUser(context.Background(), createdUser.Username)

	require.NoError(t, err)
	require.NotEmpty(t, dbRecord)

	require.Equal(t, createdUser.Username, dbRecord.Username)
	require.Equal(t, createdUser.HashedPassword, dbRecord.HashedPassword)
	require.Equal(t, createdUser.FullName, dbRecord.FullName)
	require.Equal(t, createdUser.Email, dbRecord.Email)
	require.WithinDuration(t, createdUser.CreatedAt, dbRecord.CreatedAt, time.Second)
	require.WithinDuration(t, createdUser.PasswordChangedAt, dbRecord.PasswordChangedAt, time.Second)
}

func TestUpdateUser(t *testing.T) {
	// full name update test
	oldUser := createRandomUser(t)
	newFullName := util.RandomOwner()

	dbRecord, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEmpty(t, dbRecord)

	require.Equal(t, oldUser.Username, dbRecord.Username)
	require.NotEqual(t, oldUser.FullName, dbRecord.FullName)
	require.Equal(t, oldUser.Email, dbRecord.Email)
	require.Equal(t, oldUser.HashedPassword, dbRecord.HashedPassword)

	// email update test
	oldUser = createRandomUser(t)
	newEmail := util.RandomEmail()

	dbRecord, err = testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEmpty(t, dbRecord)

	require.Equal(t, oldUser.Username, dbRecord.Username)
	require.Equal(t, oldUser.FullName, dbRecord.FullName)
	require.NotEqual(t, oldUser.Email, dbRecord.Email)
	require.Equal(t, oldUser.HashedPassword, dbRecord.HashedPassword)

	// password update test
	oldUser = createRandomUser(t)
	newPassword := util.RandomString(6)

	dbRecord, err = testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		HashedPassword: pgtype.Text{
			String: newPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEmpty(t, dbRecord)

	require.Equal(t, oldUser.Username, dbRecord.Username)
	require.Equal(t, oldUser.FullName, dbRecord.FullName)
	require.Equal(t, oldUser.Email, dbRecord.Email)
	require.NotEqual(t, oldUser.HashedPassword, dbRecord.HashedPassword)

	// all fields update test
	oldUser = createRandomUser(t)
	newFullName = util.RandomOwner()
	newEmail = util.RandomEmail()
	newPassword = util.RandomString(6)

	dbRecord, err = testStore.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
		HashedPassword: pgtype.Text{
			String: newPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.Equal(t, oldUser.Username, dbRecord.Username)

	require.NotEqual(t, oldUser.FullName, dbRecord.FullName)
	require.Equal(t, newFullName, dbRecord.FullName)
	require.NotEqual(t, oldUser.Email, dbRecord.Email)
	require.Equal(t, newEmail, dbRecord.Email)
	require.NotEqual(t, oldUser.HashedPassword, dbRecord.HashedPassword)
	require.Equal(t, newPassword, dbRecord.HashedPassword)
}
