package transfers

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

func (h *TransferHandler) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.CreateTransferResponse, error) {
	authPayload, err := core.AuthorizeUser(h.Server.TokenMaker, ctx, []string{util.DepositorRole, util.BankerRole})
	if err != nil {
		return nil, core.UnauthenticatedError(err)
	}

	// Fetch sender account
	fromAccount, err := h.validAccount(ctx, req.FromAccountId, req.Currency.String())
	if err != nil {
		return nil, err
	}

	// Check if the request is from the owner of the sender account
	if fromAccount.Owner != authPayload.Username {
		return nil, status.Errorf(codes.PermissionDenied, "from account does not belong to the authenticated user")
	}

	_, err = h.validAccount(ctx, req.ToAccountId, req.Currency.String())
	if err != nil {
		return nil, err
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountId,
		ToAccountID:   req.ToAccountId,
		Amount:        req.Amount,
		Description:   req.GetDescription(),
	}

	result, err := h.Server.Store.TransferTx(ctx, arg)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create transfer: %v", err)
	}

	response := &pb.CreateTransferResponse{
		Transfer:    result.Transfer.ToResponse(),
		FromAccount: result.FromAccount.ToResponse(),
		ToAccount:   result.ToAccount.ToResponse(),
		FromEntry:   result.FromEntry.ToResponse(),
		ToEntry:     result.ToEntry.ToResponse(),
	}

	return response, nil
}

func (h *TransferHandler) validAccount(ctx context.Context, accountId int64, currency string) (*db.Account, error) {
	account, err := h.Server.Store.GetAccount(ctx, accountId)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "account not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to fetch account details")
	}

	if account.Currency != currency {
		return nil, status.Errorf(codes.InvalidArgument, "account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
	}

	return &account, nil
}
