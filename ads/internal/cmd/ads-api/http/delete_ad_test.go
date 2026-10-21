package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adsdao "github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/dao"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/domain"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/ads/internal/cmd/ads-api/usecase"
	"github.mpi-internal.com/leboncoin/backend-hiring-assessment/common/middleware"
)

func newDeleteAdServer(fn usecase.DeleteAdFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("DELETE /ads/{adID}", middleware.RequireAuth(handlerDeleteAd(fn)))
	return httptest.NewServer(mux)
}

func deleteRequest(t *testing.T, url, authHeader string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func decodeError(t *testing.T, resp *http.Response) string {
	t.Helper()

	var body struct {
		Error string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	return body.Error
}

func TestHandlerDeleteAd(t *testing.T) {
	t.Run("returns 204 and passes the ad and user IDs to the usecase", func(t *testing.T) {
		var (
			receivedAdID   domain.AdID
			receivedUserID domain.UserID
		)
		srv := newDeleteAdServer(func(_ context.Context, id domain.AdID, userID domain.UserID) error {
			receivedAdID = id
			receivedUserID = userID
			return nil
		})
		defer srv.Close()

		resp := deleteRequest(t, srv.URL+"/ads/42", "user-1")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Equal(t, domain.AdID(42), receivedAdID)
		assert.Equal(t, domain.UserID("user-1"), receivedUserID)
	})

	t.Run("returns 401 without Authorization header", func(t *testing.T) {
		srv := newDeleteAdServer(func(_ context.Context, _ domain.AdID, _ domain.UserID) error {
			t.Fatal("usecase should not be called")
			return nil
		})
		defer srv.Close()

		resp := deleteRequest(t, srv.URL+"/ads/42", "")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 400 when adID is not a number", func(t *testing.T) {
		srv := newDeleteAdServer(func(_ context.Context, _ domain.AdID, _ domain.UserID) error {
			t.Fatal("usecase should not be called")
			return nil
		})
		defer srv.Close()

		resp := deleteRequest(t, srv.URL+"/ads/not-a-number", "user-1")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, `ad id "not-a-number" is not a number`, decodeError(t, resp))
	})

	t.Run("returns 404 when ad does not exist", func(t *testing.T) {
		srv := newDeleteAdServer(func(_ context.Context, _ domain.AdID, _ domain.UserID) error {
			return adsdao.ErrNotFound
		})
		defer srv.Close()

		resp := deleteRequest(t, srv.URL+"/ads/99", "user-1")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, "ad 99 not found", decodeError(t, resp))
	})

	t.Run("returns 403 when the ad belongs to another user", func(t *testing.T) {
		srv := newDeleteAdServer(func(_ context.Context, _ domain.AdID, _ domain.UserID) error {
			return usecase.ErrForbidden
		})
		defer srv.Close()

		resp := deleteRequest(t, srv.URL+"/ads/42", "user-2")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "ad 42 does not belong to the user", decodeError(t, resp))
	})

	t.Run("returns 500 on unexpected usecase error", func(t *testing.T) {
		srv := newDeleteAdServer(func(_ context.Context, _ domain.AdID, _ domain.UserID) error {
			return errors.New("db down")
		})
		defer srv.Close()

		resp := deleteRequest(t, srv.URL+"/ads/1", "user-1")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, "internal error", decodeError(t, resp))
	})
}
