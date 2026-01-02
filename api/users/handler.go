package users

import (
	"simplebank/api/core"
)

type UserHandler struct {
	Server *core.Server
}

func NewUserHandler(server *core.Server) *UserHandler {
	return &UserHandler{
		Server: server,
	}
}
