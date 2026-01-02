package accounts

import (
	"context"
	"errors"
	"simplebank/api/core"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AccountHandler) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.Account, error) {
	authPayload, err := core.AuthorizeUser(h.Server.TokenMaker, ctx, []string{util.DepositorRole, util.BankerRole})
	if err != nil {
		return nil, core.UnauthenticatedError(err)
	}

	account, err := h.Server.Store.GetAccount(ctx, req.GetAccountId())
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to fetch account details")
	}

	if account.Owner != authPayload.Username {
		return nil, status.Errorf(codes.PermissionDenied, "invalid account owner")
	}

	return account.ToResponse(), nil
}
