package gapi

import (
	"context"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	"simplebank/val"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.DepositorRole, util.BankerRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateListTransactionsRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
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
	accounts, err := server.store.ListAccounts(ctx, db.ListAccountsParams{
		Owner:  username,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list accounts: %v", err)
	}

	// List transfers from owned accounts
	transfers := findTransfersForAccounts(ctx, server.store, accounts)

	// Convert transfers to responses
	transactions := make([]*pb.Transaction, len(transfers))
	for i, transfer := range transfers {
		transactions[i] = transfer.ToTransaction()
	}

	return &pb.ListTransactionsResponse{
		Pagination: &pb.Pagination{
			// TODO cursor
			Count: int64(len(transfers)),
		},
		Data: transactions,
	}, nil
}

func validateListTransactionsRequest(req *pb.ListTransactionsRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if req.Username != nil {
		if err := val.ValidateUsername(req.GetUsername()); err != nil {
			violations = append(violations, fieldViolation("username", err))
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

		transfers, err := store.ListTransfers(ctx, arg)
		if err != nil {
			continue
		}
		results = append(results, transfers...)
	}

	return results
}
