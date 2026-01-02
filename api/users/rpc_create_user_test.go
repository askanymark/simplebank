package users

import (
	"context"
	"fmt"
	"reflect"
	"simplebank/api/testutil"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	"simplebank/worker"
	mockwk "simplebank/worker/mock"
	"testing"

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

func TestCreateUser(t *testing.T) {
	user, password := testutil.RandomUser(t)

	testCases := []struct {
		name          string
		body          *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore, distributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, res *pb.User, err error)
	}{
		{
			"Created",
			&pb.CreateUserRequest{
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
			func(t *testing.T, res *pb.User, err error) {
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
			coreServer := testutil.NewTestServer(t, store, taskDistributor)
			handler := NewUserHandler(coreServer)
			res, err := handler.CreateUser(context.Background(), tc.body)
			tc.checkResponse(t, res, err)
		})
	}
}
