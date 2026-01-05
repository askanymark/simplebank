package api

import (
	"fmt"
	"simplebank/api/accounts"
	"simplebank/api/core"
	"simplebank/api/transfers"
	"simplebank/api/users"
	db "simplebank/db/sqlc"
	"simplebank/token"
	"simplebank/util"
	"simplebank/worker"
)

type Server struct {
	*core.Core
	*accounts.AccountHandler
	*users.UserHandler
	*transfers.TransferHandler
}

func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker([]byte(config.TokenSymmetricKey))
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	apiCore := &core.Core{
		Store:           store,
		TokenMaker:      tokenMaker,
		Config:          config,
		TaskDistributor: taskDistributor,
	}

	server := &Server{
		Core:            apiCore,
		AccountHandler:  accounts.NewAccountHandler(apiCore),
		UserHandler:     users.NewUserHandler(apiCore),
		TransferHandler: transfers.NewTransferHandler(apiCore),
	}

	return server, nil
}
