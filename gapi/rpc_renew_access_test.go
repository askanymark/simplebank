package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	mockwk "simplebank/worker/mock"
)

func TestRenewAccess(t *testing.T) {
	user, _ := randomUser(t)

	role := util.DepositorRole
	duration := time.Minute

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore, server *Server) (string, time.Time)
		checkResponse func(t *testing.T, res *pb.RenewAccessResponse, err error)
	}{
		{
			"OK",
			func(store *mockdb.MockStore, server *Server) (string, time.Time) {
				refreshToken, payload, err := server.tokenMaker.CreateToken(user.Username, role, duration)
				require.NoError(t, err)

				session := db.Session{
					ID:           payload.ID,
					Username:     user.Username,
					RefreshToken: refreshToken,
					IsBlocked:    false,
					ExpiresAt: pgtype.Timestamp{
						Time:  payload.ExpiredAt,
						Valid: true,
					},
				}

				store.EXPECT().
					GetSession(gomock.Any(), gomock.Eq(payload.ID)).
					Times(1).
					Return(session, nil)

				return refreshToken, payload.ExpiredAt
			},
			func(t *testing.T, res *pb.RenewAccessResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res.AccessToken)
				require.NotEmpty(t, res.AccessTokenExpiresAt)
			},
		},
		{
			"Unauthenticated",
			func(store *mockdb.MockStore, server *Server) (string, time.Time) {
				return "invalid-token", time.Time{}
			},
			func(t *testing.T, res *pb.RenewAccessResponse, err error) {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			"SessionNotFound",
			func(store *mockdb.MockStore, server *Server) (string, time.Time) {
				refreshToken, payload, err := server.tokenMaker.CreateToken(user.Username, role, duration)
				require.NoError(t, err)

				store.EXPECT().
					GetSession(gomock.Any(), gomock.Eq(payload.ID)).
					Times(1).
					Return(db.Session{}, db.ErrRecordNotFound)

				return refreshToken, payload.ExpiredAt
			},
			func(t *testing.T, res *pb.RenewAccessResponse, err error) {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			"SessionBlocked",
			func(store *mockdb.MockStore, server *Server) (string, time.Time) {
				refreshToken, payload, err := server.tokenMaker.CreateToken(user.Username, role, duration)
				require.NoError(t, err)

				session := db.Session{
					ID:           payload.ID,
					Username:     user.Username,
					RefreshToken: refreshToken,
					IsBlocked:    true,
					ExpiresAt: pgtype.Timestamp{
						Time:  payload.ExpiredAt,
						Valid: true,
					},
				}

				store.EXPECT().
					GetSession(gomock.Any(), gomock.Eq(payload.ID)).
					Times(1).
					Return(session, nil)

				return refreshToken, payload.ExpiredAt
			},
			func(t *testing.T, res *pb.RenewAccessResponse, err error) {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"UserMismatch",
			func(store *mockdb.MockStore, server *Server) (string, time.Time) {
				refreshToken, payload, err := server.tokenMaker.CreateToken(user.Username, role, duration)
				require.NoError(t, err)

				session := db.Session{
					ID:           payload.ID,
					Username:     "other-user",
					RefreshToken: refreshToken,
					IsBlocked:    false,
					ExpiresAt: pgtype.Timestamp{
						Time:  payload.ExpiredAt,
						Valid: true,
					},
				}

				store.EXPECT().
					GetSession(gomock.Any(), gomock.Eq(payload.ID)).
					Times(1).
					Return(session, nil)

				return refreshToken, payload.ExpiredAt
			},
			func(t *testing.T, res *pb.RenewAccessResponse, err error) {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"TokenMismatch",
			func(store *mockdb.MockStore, server *Server) (string, time.Time) {
				refreshToken, payload, err := server.tokenMaker.CreateToken(user.Username, role, duration)
				require.NoError(t, err)

				session := db.Session{
					ID:           payload.ID,
					Username:     user.Username,
					RefreshToken: "mismatch-token",
					IsBlocked:    false,
					ExpiresAt: pgtype.Timestamp{
						Time:  payload.ExpiredAt,
						Valid: true,
					},
				}

				store.EXPECT().
					GetSession(gomock.Any(), gomock.Eq(payload.ID)).
					Times(1).
					Return(session, nil)

				return refreshToken, payload.ExpiredAt
			},
			func(t *testing.T, res *pb.RenewAccessResponse, err error) {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"SessionExpired",
			func(store *mockdb.MockStore, server *Server) (string, time.Time) {
				refreshToken, payload, err := server.tokenMaker.CreateToken(user.Username, role, duration)
				require.NoError(t, err)

				session := db.Session{
					ID:           payload.ID,
					Username:     user.Username,
					RefreshToken: refreshToken,
					IsBlocked:    false,
					ExpiresAt: pgtype.Timestamp{
						Time:  time.Now().Add(-time.Minute),
						Valid: true,
					},
				}

				store.EXPECT().
					GetSession(gomock.Any(), gomock.Eq(payload.ID)).
					Times(1).
					Return(session, nil)

				return refreshToken, payload.ExpiredAt
			},
			func(t *testing.T, res *pb.RenewAccessResponse, err error) {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			"InternalError",
			func(store *mockdb.MockStore, server *Server) (string, time.Time) {
				refreshToken, payload, err := server.tokenMaker.CreateToken(user.Username, role, duration)
				require.NoError(t, err)

				store.EXPECT().
					GetSession(gomock.Any(), gomock.Eq(payload.ID)).
					Times(1).
					Return(db.Session{}, sql.ErrConnDone)

				return refreshToken, payload.ExpiredAt
			},
			func(t *testing.T, res *pb.RenewAccessResponse, err error) {
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

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			server := newTestServer(t, store, taskDistributor)
			refreshToken, _ := tc.buildStubs(store, server)

			req := &pb.RenewAccessRequest{
				RefreshToken: refreshToken,
			}

			res, err := server.RenewAccess(context.Background(), req)
			tc.checkResponse(t, res, err)
		})
	}
}
