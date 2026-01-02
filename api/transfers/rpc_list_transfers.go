package transfers

import (
	"context"
	"simplebank/api/core"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	"simplebank/val"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *TransferHandler) ListTransfers(ctx context.Context, req *pb.ListTransfersRequest) (*pb.ListTransfersResponse, error) {
	authPayload, err := core.AuthorizeUser(h.Server.TokenMaker, ctx, []string{util.DepositorRole, util.BankerRole})
	if err != nil {
		return nil, core.UnauthenticatedError(err)
	}

	violations := validateListTransfersRequest(req)
	if violations != nil {
		return nil, core.InvalidArgumentError(violations)
	}

	// Only bankers can list transactions of other users
	if req.Username != nil && !util.IsBanker(authPayload.Role) {
		return nil, status.Errorf(codes.PermissionDenied, "cannot list transactions of other users")
	}

	username := authPayload.Username
	if req.Username != nil {
		username = req.GetUsername()
	}

	// Find accounts the user owns
	accounts, err := h.Server.Store.ListAccounts(ctx, db.ListAccountsParams{
		Owner:  username,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list accounts: %v", err)
	}

	// List transfers from owned accounts
	transfers := findTransfersForAccounts(ctx, h.Server.Store, accounts)

	response := &pb.ListTransfersResponse{
		Pagination: &pb.Pagination{
			// TODO cursor
			Count: int64(len(transfers)),
		},
		Data: make([]*pb.Transfer, len(transfers)),
	}

	for i, transfer := range transfers {
		response.Data[i] = transfer.ToResponse()
	}

	return response, nil
}

func validateListTransfersRequest(req *pb.ListTransfersRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if req.Username != nil {
		if err := val.ValidateUsername(req.GetUsername()); err != nil {
			violations = append(violations, core.FieldViolation("username", err))
		}
	}

	return violations
}

func findTransfersForAccounts(ctx context.Context, store db.Store, accounts []db.Account) []db.Transfer {
	var results []db.Transfer

	for _, account := range accounts {
		arg := db.ListTransfersParams{
			FromAccountID: account.ID,
			ToAccountID:   account.ID,
			Limit:         10,
			Offset:        0,
		}

		listTransfers, err := store.ListTransfers(ctx, arg)
		if err != nil {
			continue
		}
		results = append(results, listTransfers...)
	}

	return results
}
