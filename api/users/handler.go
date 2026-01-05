package users

import (
	"simplebank/api/core"
)

type UserHandler struct {
	core *core.Core
}

func NewUserHandler(core *core.Core) *UserHandler {
	return &UserHandler{
		core: core,
	}
}
