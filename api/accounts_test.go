package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/util"
	"strconv"
	"testing"
	"time"
)

func TestGetAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountId     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"OK",
			account.ID,
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			"NotFound",
			account.ID,
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			"InternalError",
			account.ID,
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			"InvalidId",
			0,
			func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			store := mockdb.NewMockStore(controller)
			tc.buildStubs(store)

			// start the server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%d", tc.accountId)

			request, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	var account db.Account
	var body []byte

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"Created",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				body = []byte(fmt.Sprintf(`{"owner":"%s","currency":"%s"}`, account.Owner, account.Currency))

				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
						Owner:    account.Owner,
						Balance:  0,
						Currency: account.Currency,
					})).
					Times(1).
					Return(db.Account{
						ID:        account.ID,
						Owner:     account.Owner,
						Balance:   0,
						Currency:  account.Currency,
						CreatedAt: account.CreatedAt,
					}, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)

				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				var actual db.Account
				err = json.Unmarshal(data, &actual)
				require.NoError(t, err)
				require.NotZero(t, actual.ID)
				require.Equal(t, account.Owner, actual.Owner)
				require.Equal(t, account.Currency, actual.Currency)
				require.Equal(t, int64(0), actual.Balance)
				require.NotZero(t, actual.CreatedAt)
			},
		},
		{
			"BadRequest",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				body = []byte(fmt.Sprintf(`{"owner":123,"currency":"beef mince"}`))

				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			"InternalError",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				body = []byte(fmt.Sprintf(`{"owner":"%s","currency":"%s"}`, account.Owner, account.Currency))

				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
						Owner:    account.Owner,
						Balance:  0,
						Currency: account.Currency,
					})).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			store := mockdb.NewMockStore(controller)
			tc.buildStubs(store)

			// start the server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := "/accounts"

			request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
			request.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccounts(t *testing.T) {
	var query string

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"OK",
			func(store *mockdb.MockStore) {
				query = "?page_id=1&page_size=5"

				store.EXPECT().ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				})).Times(1).Return([]db.Account{}, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			"InvalidQuery",
			func(store *mockdb.MockStore) {
				query = "?this_will_fail"

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			"InternalError",
			func(store *mockdb.MockStore) {
				query = "?page_id=1&page_size=5"

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
						Limit:  5,
						Offset: 0,
					})).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			store := mockdb.NewMockStore(controller)
			tc.buildStubs(store)

			// start the server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest("GET", fmt.Sprintf("/accounts%s", query), nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateAccounts(t *testing.T) {
	var account db.Account
	var body []byte
	var accountId string

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"NoContent",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)
				body = []byte(fmt.Sprintf(`{"balance":%d}`, int64(10)))

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)

				store.EXPECT().UpdateAccount(gomock.Any(), gomock.Eq(db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance + 10,
				})).Times(1).Return(account, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			"InvalidId",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = "eee"
				body = []byte(fmt.Sprintf(`{"balance":%d}`, 10))

				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().UpdateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			"InvalidBody",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)
				body = []byte("this will fail")
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			"NotFound",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)
				body = []byte(fmt.Sprintf(`{"balance":%d}`, int64(10)))

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().UpdateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			"InternalError during select",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)
				body = []byte(fmt.Sprintf(`{"balance":%d}`, int64(10)))

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
				store.EXPECT().UpdateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			"InternalError during update",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)
				body = []byte(fmt.Sprintf(`{"balance":%d}`, int64(10)))

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
				store.EXPECT().UpdateAccount(gomock.Any(), gomock.Eq(db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance + 10,
				})).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			store := mockdb.NewMockStore(controller)
			tc.buildStubs(store)

			// start the server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%s", accountId)

			request, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
			request.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	var account db.Account
	var accountId string

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"NoContent",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			"InvalidId",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = "this will fail"

				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			"NotFound",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			"InternalError during select",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			"InternalError during delete",
			func(store *mockdb.MockStore) {
				account = randomAccount()
				accountId = strconv.FormatInt(account.ID, 10)

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
				store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(sql.ErrConnDone)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			store := mockdb.NewMockStore(controller)

			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%s", accountId)
			request, err := http.NewRequest("DELETE", url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:        util.RandomInt(1, 1000),
		Owner:     util.RandomOwner(),
		Balance:   util.RandomMoney(),
		Currency:  util.RandomCurrency(),
		CreatedAt: time.Now().UTC(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var actual db.Account
	err = json.Unmarshal(data, &actual)
	require.NoError(t, err)
	require.Equal(t, account, actual)
}
