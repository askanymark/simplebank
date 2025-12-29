package gapi

import (
	"context"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/token"
	"simplebank/util"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListTransactions(t *testing.T) {
	user, _ := randomUser(t)
	banker, _ := randomUser(t)
	banker.Role = util.BankerRole

	otherUser, _ := randomUser(t)

	account1 := randomAccount(user.Username)
	account2 := randomAccount(user.Username)

	transfer1 := randomTransfer(account1.ID, account2.ID)
	transfer2 := randomTransfer(account2.ID, account1.ID)

	testCases := []struct {
		name          string
		req           *pb.ListTransactionsRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.ListTransactionsResponse, err error)
	}{
		{
			"OK",
			&pb.ListTransactionsRequest{},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  10,
						Offset: 0,
					})).
					Times(1).
					Return([]db.Account{account1, account2}, nil)

				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Eq(db.ListTransfersParams{
						FromAccountID: account1.ID,
						ToAccountID:   account1.ID,
						Limit:         10,
						Offset:        0,
					})).
					Times(1).
					Return([]db.Transfer{transfer1, transfer2}, nil)

				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Eq(db.ListTransfersParams{
						FromAccountID: account2.ID,
						ToAccountID:   account2.ID,
						Limit:         10,
						Offset:        0,
					})).
					Times(1).
					Return([]db.Transfer{transfer1, transfer2}, nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.ListTransactionsResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Len(t, res.Data, 4)
			},
		},
		{
			"BankerListOtherUser",
			&pb.ListTransactionsRequest{
				Username: &otherUser.Username,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
						Owner:  otherUser.Username,
						Limit:  10,
						Offset: 0,
					})).
					Times(1).
					Return([]db.Account{account1}, nil)

				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Transfer{transfer1}, nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				// Bankers must also have DepositorRole to pass authorizeUser check if it's required
				// But wait, ListTransactions only allows DepositorRole in authorizeUser!
				return newContextWithBearerToken(t, tokenMaker, banker, time.Minute)
			},
			func(t *testing.T, res *pb.ListTransactionsResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Len(t, res.Data, 1)
			},
		},
		{
			"PermissionDenied",
			&pb.ListTransactionsRequest{
				Username: &otherUser.Username,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.ListTransactionsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"Unauthenticated",
			&pb.ListTransactionsRequest{},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			func(t *testing.T, res *pb.ListTransactionsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			"InternalError",
			&pb.ListTransactionsRequest{},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, context.DeadlineExceeded)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.ListTransactionsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.ListTransactions(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func randomTransfer(fromAccountID, toAccountID int64) db.Transfer {
	return db.Transfer{
		ID:            util.RandomInt(1, 1000),
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        util.RandomMoney(),
		CreatedAt:     time.Now().UTC(),
		Description: pgtype.Text{
			String: util.RandomString(18),
			Valid:  true,
		},
	}
}
