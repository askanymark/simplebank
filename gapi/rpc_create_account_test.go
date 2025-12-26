package gapi

import (
	"context"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/token"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServer_CreateAccount(t *testing.T) {
	user, _ := randomUser(t)
	expectedCurrency := pb.Currency_GBP
	currencyPtr := expectedCurrency.Enum()

	testCases := []struct {
		name          string
		body          *pb.CreateAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.Account, err error)
	}{
		{
			"Created",
			&pb.CreateAccountRequest{
				Currency: expectedCurrency,
			},
			func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Balance:  0,
					Currency: currencyPtr.String(),
				}

				newAccount := db.Account{
					ID:        1,
					Owner:     user.FullName,
					Balance:   0,
					Currency:  currencyPtr.String(),
					CreatedAt: time.Now().UTC(),
				}

				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(newAccount, nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user, time.Minute)
			},
			func(t *testing.T, res *pb.Account, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.FullName, res.Owner)
				require.Equal(t, expectedCurrency, res.Currency)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			tc.buildStubs(store)

			// start the server and send the request
			server := newTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.CreateAccount(ctx, tc.body)
			tc.checkResponse(t, res, err)
		})
	}
}
