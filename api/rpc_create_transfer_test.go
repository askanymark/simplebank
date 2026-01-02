package api

import (
	"context"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/token"
	"simplebank/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateTransfer(t *testing.T) {
	user, _ := randomUser(t)
	account1 := randomAccount(user.Username)
	account2 := randomAccount(util.RandomOwner())

	account1.Currency = util.USD
	account2.Currency = util.USD

	amount := int64(10)

	testCases := []struct {
		name          string
		req           *pb.CreateTransferRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.CreateTransferResponse, err error)
	}{
		{
			"OK",
			&pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
				Currency:      pb.Currency_USD,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)

				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.TransferTxResult{
						Transfer:    db.Transfer{ID: 1, FromAccountID: account1.ID, ToAccountID: account2.ID, Amount: amount},
						FromAccount: account1,
						ToAccount:   account2,
						FromEntry:   db.Entry{AccountID: account1.ID, Amount: -amount},
						ToEntry:     db.Entry{AccountID: account2.ID, Amount: amount},
					}, nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
			},
		},
		{
			"Unauthorized",
			&pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
				Currency:      pb.Currency_USD,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			"FromAccountNotFound",
			&pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
				Currency:      pb.Currency_USD,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(db.Account{}, db.ErrRecordNotFound)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			"ToAccountNotFound",
			&pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
				Currency:      pb.Currency_USD,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(db.Account{}, db.ErrRecordNotFound)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			"FromAccountCurrencyMismatch",
			&pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
				Currency:      pb.Currency_EUR,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			"ToAccountCurrencyMismatch",
			&pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
				Currency:      pb.Currency_USD,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				account2Mismatch := account2
				account2Mismatch.Currency = util.EUR
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2Mismatch, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			"PermissionDenied",
			&pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
				Currency:      pb.Currency_USD,
			},
			func(store *mockdb.MockStore) {
				otherAccount := account1
				otherAccount.Owner = "other_owner"
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(otherAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"InternalErrorTransferTx",
			&pb.CreateTransferRequest{
				FromAccountId: account1.ID,
				ToAccountId:   account2.ID,
				Amount:        amount,
				Currency:      pb.Currency_USD,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(account1, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.TransferTxResult{}, context.DeadlineExceeded)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.CreateTransfer(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
