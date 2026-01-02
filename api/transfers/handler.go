package transfers

import (
	"simplebank/api/core"
)

type TransferHandler struct {
	Server *core.Server
}

func NewTransferHandler(server *core.Server) *TransferHandler {
	return &TransferHandler{
		Server: server,
	}
}
