package api

import (
	"context"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/pb/accounts"
	"simplebank/util"
	"simplebank/val"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ListAccounts(ctx context.Context, req *accounts.ListAccountsRequest) (*accounts.ListAccountsResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.DepositorRole, util.BankerRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateListAccountsRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	// Only bankers can list accounts of other users
	if req.Username != nil && !util.IsBanker(authPayload.Role) {
		return nil, status.Errorf(codes.PermissionDenied, "cannot list accounts of other users")
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

	response := *accounts.ListAccountsResponse{
		Pagination: &pb.Pagination{
			// TODO cursor
			Count: int64(len(accounts)),
		},
		Data: make([]*accounts.Account, len(accounts)),
	}

	for i, account := range accounts {
		response.Data[i] = account.ToResponse()
	}

	return response, nil
}

func validateListAccountsRequest(req *accounts.ListAccountsRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if req.Username != nil {
		if err := val.ValidateUsername(req.GetUsername()); err != nil {
			violations = append(violations, fieldViolation("username", err))
		}
	}

	return violations
}
