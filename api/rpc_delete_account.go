package api

import (
	"context"
	"errors"
	db "simplebank/db/sqlc"
	"simplebank/pb/accounts"
	"simplebank/util"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) DeleteAccount(ctx context.Context, req *accounts.DeleteAccountRequest) (*emptypb.Empty, error) {
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

	// Only bankers can delete accounts they don't own
	if account.Owner != authPayload.Username && !util.IsBanker(authPayload.Role) {
		return nil, status.Errorf(codes.PermissionDenied, "cannot delete other user's account")
	}

	if err = server.store.DeleteAccount(ctx, req.GetAccountId()); err != nil {
		errCode := db.ErrorCode(err)

		if errCode == db.ForeignKeyViolation {
			return nil, status.Errorf(codes.FailedPrecondition, "cannot delete account with transactions")
		}

		return nil, status.Errorf(codes.Internal, "failed to delete account: %s", err)
	}

	return &emptypb.Empty{}, nil
}
