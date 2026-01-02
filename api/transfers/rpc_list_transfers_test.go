package transfers

import (
	"context"
	"simplebank/api/testutil"
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

func TestListTransactions(t *testing.T) {
	user, _ := testutil.RandomUser(t)
	banker, _ := testutil.RandomUser(t)
	banker.Role = util.BankerRole

	otherUser, _ := testutil.RandomUser(t)

	account1 := testutil.RandomAccount(user.Username)
	account2 := testutil.RandomAccount(user.Username)

	transfer1 := testutil.RandomTransfer(account1.ID, account2.ID)
	transfer2 := testutil.RandomTransfer(account2.ID, account1.ID)

	testCases := []struct {
		name          string
		req           *pb.ListTransfersRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.ListTransfersResponse, err error)
	}{
		{
			"OK",
			&pb.ListTransfersRequest{},
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
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, res *pb.ListTransfersResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Len(t, res.Data, 4)
			},
		},
		{
			"BankerListOtherUser",
			&pb.ListTransfersRequest{
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
				return testutil.NewContextWithBearerToken(t, tokenMaker, banker.Username, banker.Role, time.Minute)
			},
			func(t *testing.T, res *pb.ListTransfersResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Len(t, res.Data, 1)
			},
		},
		{
			"PermissionDenied",
			&pb.ListTransfersRequest{
				Username: &otherUser.Username,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, res *pb.ListTransfersResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"Unauthenticated",
			&pb.ListTransfersRequest{},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			func(t *testing.T, res *pb.ListTransfersResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			"InternalError",
			&pb.ListTransfersRequest{},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, context.DeadlineExceeded)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, res *pb.ListTransfersResponse, err error) {
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

			coreServer := testutil.NewTestServer(t, store, nil)
			handler := NewTransferHandler(coreServer)
			ctx := tc.buildContext(t, coreServer.TokenMaker)
			res, err := handler.ListTransfers(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
