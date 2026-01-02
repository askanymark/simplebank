package core

import (
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/token"
	"simplebank/util"
	"simplebank/worker"
)

type Server struct {
	pb.UnimplementedSimplebankServer
	Config          util.Config
	Store           db.Store
	TokenMaker      token.Maker
	TaskDistributor worker.TaskDistributor
}
