package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"simplebank/token"
	"simplebank/util"
	"testing"
	"time"
)

func addAuthorization(t *testing.T, request *http.Request, maker token.Maker, authorizationType, username string, duration time.Duration) {
	createdToken, payload, err := maker.CreateToken(username, util.DepositorRole, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, createdToken)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, maker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"OK",
			func(t *testing.T, request *http.Request, maker token.Maker) {
				addAuthorization(t, request, maker, authorizationTypeBearer, "user", time.Minute)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			"NoAuthorization",
			func(t *testing.T, request *http.Request, maker token.Maker) {

			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			"UnsupportedAuthorization",
			func(t *testing.T, request *http.Request, maker token.Maker) {
				addAuthorization(t, request, maker, "unsupported", "user", time.Minute)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			"InvalidAuthorizationFormat",
			func(t *testing.T, request *http.Request, maker token.Maker) {
				addAuthorization(t, request, maker, "", "user", time.Minute)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			"ExpiredToken",
			func(t *testing.T, request *http.Request, maker token.Maker) {
				addAuthorization(t, request, maker, authorizationTypeBearer, "user", -time.Minute)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)
			authPath := "/auth"
			server.router.GET(authPath, authMiddleware(server.tokenMaker), func(context *gin.Context) {
				context.JSON(http.StatusOK, gin.H{})
			})

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
