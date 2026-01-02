package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"simplebank/api/core"
	db "simplebank/db/sqlc"
	"simplebank/token"
	"simplebank/util"
	"simplebank/worker"
)

func NewTestServer(t *testing.T, store db.Store, distributor worker.TaskDistributor) *core.Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	tokenMaker, err := token.NewPasetoMaker([]byte(config.TokenSymmetricKey))
	require.NoError(t, err)

	return &core.Server{
		Config:          config,
		Store:           store,
		TokenMaker:      tokenMaker,
		TaskDistributor: distributor,
	}
}

func NewContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, role string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", core.BearerPrefix, accessToken)

	md := metadata.MD{
		core.AuthorizationHeader: []string{
			bearerToken,
		},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}

func RandomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		Role:           util.DepositorRole,
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func RandomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func RandomTransfer(fromAccountID, toAccountID int64) db.Transfer {
	return db.Transfer{
		ID:            util.RandomInt(1, 1000),
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        util.RandomMoney(),
	}
}
