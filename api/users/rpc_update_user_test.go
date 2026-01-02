package users

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

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUpdateUserAPI(t *testing.T) {
	user, _ := testutil.RandomUser(t)

	newName := util.RandomOwner()
	newEmail := util.RandomEmail()

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, resp *pb.User, err error)
	}{
		{
			"OK",
			&pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: pgtype.Text{
						String: newName,
						Valid:  true,
					},
					Email: pgtype.Text{
						String: newEmail,
						Valid:  true,
					},
				}

				updatedUser := db.User{
					Username:       user.Username,
					HashedPassword: user.HashedPassword,
					FullName:       newName,
					Email:          newEmail,
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedUser, nil)
			},
			func(t *testing.T, tokenMaker token.Maker) context.Context {
				return testutil.NewContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute)
			},
			func(t *testing.T, res *pb.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.Username)
				require.Equal(t, newName, res.FullName)
				require.Equal(t, newEmail, res.Email)
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
			handler := NewUserHandler(coreServer)

			ctx := tc.buildContext(t, coreServer.TokenMaker)
			res, err := handler.UpdateUser(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
