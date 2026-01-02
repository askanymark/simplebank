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
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *AccountHandler) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*emptypb.Empty, error) {
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

	// Only bankers can delete accounts they don't own
	if account.Owner != authPayload.Username && !util.IsBanker(authPayload.Role) {
		return nil, status.Errorf(codes.PermissionDenied, "cannot delete other user's account")
	}

	if err = h.Server.Store.DeleteAccount(ctx, req.GetAccountId()); err != nil {
		errCode := db.ErrorCode(err)

		if errCode == db.ForeignKeyViolation {
			return nil, status.Errorf(codes.FailedPrecondition, "cannot delete account with transactions")
		}

		return nil, status.Errorf(codes.Internal, "failed to delete account: %s", err)
	}

	return &emptypb.Empty{}, nil
}
