package accounts

import (
	"simplebank/api/core"
)

type AccountHandler struct {
	core *core.Core
}

func NewAccountHandler(core *core.Core) *AccountHandler {
	return &AccountHandler{
		core: core,
	}
}
