package userclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/userclient"
	usersclient "github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/users/pkg/api"
)

func newUserServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/users/{userID}", func(rw http.ResponseWriter, r *http.Request) {
		if r.PathValue("userID") != "user-1" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = rw.Write([]byte(`{"user_id":"user-1","nickname":"alice"}`))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv
}

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()
	client := userclient.New(newUserServer(t).URL, time.Second)

	t.Run("nominal", func(t *testing.T) {
		got, err := client.GetUserByID(ctx, "user-1")

		require.NoError(t, err)
		assert.Equal(t, usersclient.User{UserID: "user-1", Nickname: "alice"}, got)
	})

	t.Run("error user not found", func(t *testing.T) {
		_, err := client.GetUserByID(ctx, "unknown")

		require.Error(t, err)
		assert.ErrorIs(t, err, api.ErrNotFound)
	})
}
