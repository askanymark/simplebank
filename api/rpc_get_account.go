package api

import (
	"context"
	"errors"
	db "simplebank/db/sqlc"
	"simplebank/pb/accounts"
	"simplebank/util"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) GetAccount(ctx context.Context, req *accounts.GetAccountRequest) (*accounts.Account, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.DepositorRole, util.BankerRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	account, err := server.store.GetAccount(ctx, req.GetAccountId())
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
