package accounts

import (
	"simplebank/api/core"
)

type AccountHandler struct {
	Server *core.Server
}

func NewAccountHandler(server *core.Server) *AccountHandler {
	return &AccountHandler{
		Server: server,
	}
}
