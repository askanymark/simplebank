package api

import (
	"context"
	"fmt"
	"reflect"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/pb/users"
	"simplebank/util"
	"simplebank/worker"
	mockwk "simplebank/worker/mock"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type eqCreateUserTxParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
	user     db.User
}

func (expected *eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(expected.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}

	expected.arg.HashedPassword = actualArg.HashedPassword
	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	err = actualArg.AfterCreate(expected.user)

	return err == nil
}

func (expected *eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", expected.arg, expected.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return &eqCreateUserTxParamsMatcher{arg, password, user}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	user = db.User{
		Username:          util.RandomOwner(),
		Role:              util.DepositorRole,
		HashedPassword:    hashedPassword,
		FullName:          util.RandomOwner(),
		Email:             util.RandomEmail(),
		CreatedAt:         time.Now().UTC(),
		PasswordChangedAt: time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC),
	}

	return
}

func TestCreateUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          *users.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore, distributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, res *users.User, err error)
	}{
		{
			"Created",
			&users.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			func(store *mockdb.MockStore, distributor *mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, password, user)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)

				taskPayload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}
				distributor.EXPECT().
					DistributeSendVerifyEmailTask(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			func(t *testing.T, res *users.User, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.Username)
				require.Equal(t, user.FullName, res.FullName)
				require.Equal(t, user.Email, res.Email)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			tc.buildStubs(store, taskDistributor)

			// start the server and send request
			server := newTestServer(t, store, taskDistributor)
			res, err := server.CreateUser(context.Background(), tc.body)
			tc.checkResponse(t, res, err)
		})
	}
}
