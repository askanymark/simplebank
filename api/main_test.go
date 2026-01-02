package api

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	db "simplebank/db/sqlc"
	"simplebank/token"
	"simplebank/util"
	"simplebank/worker"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store, distributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, distributor)
	require.NoError(t, err)

	return server
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, user db.User, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(user.Username, user.Role, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", bearerPrefix, accessToken)

	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}
