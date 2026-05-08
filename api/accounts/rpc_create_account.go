package accounts

import (
	"context"
	"simplebank/api/core"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	"simplebank/val"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *AccountHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.Account, error) {
	ctx, span := otel.Tracer("accounts").Start(ctx, "CreateAccount")
	defer span.End()

	span.SetAttributes(
		attribute.String("account.currency", req.GetCurrency().String()),
	)

	authPayload, err := core.AuthorizeUser(h.core.TokenMaker, ctx, []string{util.DepositorRole})
	if err != nil {
		span.RecordError(err)
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

	account, err := h.core.Store.CreateAccount(ctx, arg)
	if err != nil {
		span.RecordError(err)
		return nil, status.Errorf(codes.Internal, "failed to create account")
	}

	span.SetAttributes(attribute.Int64("account.id", account.ID))
	return account.ToResponse(), nil
}

func validateCreateAccountRequest(req *pb.CreateAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateCurrency(req.GetCurrency().String()); err != nil {
		violations = append(violations, core.FieldViolation("currency", err))
	}

	return violations
}
