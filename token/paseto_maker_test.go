package token

import (
	"github.com/stretchr/testify/require"
	"simplebank/util"
	"testing"
	"time"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker([]byte(util.RandomString(32)))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)
	role := util.DepositorRole

	token, payload, err := maker.CreateToken(username, role, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotZero(t, payload.ID)
	require.Equal(t, role, payload.Role)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker([]byte(util.RandomString(32)))
	require.NoError(t, err)

	token, payload, err := maker.CreateToken(util.RandomOwner(), util.DepositorRole, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

func TestInvalidPasetoToken(t *testing.T) {
	payload, err := NewPayload(util.RandomOwner(), util.DepositorRole, time.Minute)
	require.NoError(t, err)

	maker, err := NewPasetoMaker([]byte(util.RandomString(32)))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(util.RandomString(6))
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
