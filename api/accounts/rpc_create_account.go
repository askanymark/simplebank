package accounts

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

func (h *AccountHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.Account, error) {
	authPayload, err := core.AuthorizeUser(h.Server.TokenMaker, ctx, []string{util.DepositorRole})
	if err != nil {
		return nil, core.UnauthenticatedError(err)
	}

	violations := validateCreateAccountRequest(req)
	if violations != nil {
		return nil, core.InvalidArgumentError(violations)
	}

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.GetCurrency().String(),
		Balance:  0,
	}

	account, err := h.Server.Store.CreateAccount(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create account")
	}

	return account.ToResponse(), nil
}

func validateCreateAccountRequest(req *pb.CreateAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateCurrency(req.GetCurrency().String()); err != nil {
		violations = append(violations, core.FieldViolation("currency", err))
	}

	return violations
}
