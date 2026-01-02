package accounts

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

func TestListAccounts(t *testing.T) {
	user, _ := testutil.RandomUser(t)
	banker, _ := testutil.RandomUser(t)
	banker.Role = util.BankerRole
	otherUser, _ := testutil.RandomUser(t)

	n := 5
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = testutil.RandomAccount(user.Username)
	}

	otherAccounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		otherAccounts[i] = testutil.RandomAccount(otherUser.Username)
	}

	testCases := []struct {
		name          string
		req           *pb.ListAccountsRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.ListAccountsResponse, err error)
	}{
		{
			"OK",
			&pb.ListAccountsRequest{
				Limit: 10,
			},
			func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  10,
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Len(t, res.Data, n)
			},
		},
		{
			"OKBanker",
			&pb.ListAccountsRequest{
				Username: &otherUser.Username,
				Limit:    10,
			},
			func(store *mockdb.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  otherUser.Username,
					Limit:  10,
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(otherAccounts, nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, banker.Username, banker.Role, time.Minute)
			},
			func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Len(t, res.Data, n)
			},
		},
		{
			"PermissionDenied",
			&pb.ListAccountsRequest{
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
			func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"Unauthorized",
			&pb.ListAccountsRequest{},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			"InternalError",
			&pb.ListAccountsRequest{},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Account{}, context.DeadlineExceeded)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			"InvalidUsername",
			&pb.ListAccountsRequest{
				Username: func() *string { s := "invalid#user"; return &s }(),
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, banker.Username, banker.Role, time.Minute)
			},
			func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
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
			handler := NewAccountHandler(coreServer)
			ctx := tc.buildContext(t, coreServer.TokenMaker)
			res, err := handler.ListAccounts(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
