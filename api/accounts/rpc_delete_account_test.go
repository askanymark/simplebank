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

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeleteAccount(t *testing.T) {
	user, _ := testutil.RandomUser(t)
	account := testutil.RandomAccount(user.Username)

	banker, _ := testutil.RandomUser(t)
	banker.Role = util.BankerRole

	testCases := []struct {
		name          string
		req           *pb.DeleteAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, err error)
	}{
		{
			"OK",
			&pb.DeleteAccountRequest{
				AccountId: account.ID,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			"OKBanker",
			&pb.DeleteAccountRequest{
				AccountId: account.ID,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, banker.Username, banker.Role, time.Minute)
			},
			func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			"NotFound",
			&pb.DeleteAccountRequest{
				AccountId: account.ID,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, db.ErrRecordNotFound)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			"InternalErrorOnGet",
			&pb.DeleteAccountRequest{
				AccountId: account.ID,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, context.DeadlineExceeded)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			"PermissionDenied",
			&pb.DeleteAccountRequest{
				AccountId: account.ID,
			},
			func(store *mockdb.MockStore) {
				otherUser, _ := testutil.RandomUser(t)
				otherAccount := testutil.RandomAccount(otherUser.Username)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(otherAccount, nil)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"InternalErrorOnDelete",
			&pb.DeleteAccountRequest{
				AccountId: account.ID,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(context.DeadlineExceeded)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			"ForeignKeyViolation",
			&pb.DeleteAccountRequest{
				AccountId: account.ID,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(&pgconn.PgError{Code: db.ForeignKeyViolation})
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.FailedPrecondition, st.Code())
			},
		},
		{
			"Unauthorized",
			&pb.DeleteAccountRequest{
				AccountId: account.ID,
			},
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			func(t *testing.T, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
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
			_, err := handler.DeleteAccount(ctx, tc.req)
			tc.checkResponse(t, err)
		})
	}
}
