package transfers

import (
	"simplebank/api/core"
)

type TransferHandler struct {
	core *core.Core
}

func NewTransferHandler(core *core.Core) *TransferHandler {
	return &TransferHandler{
		core: core,
	}
}
