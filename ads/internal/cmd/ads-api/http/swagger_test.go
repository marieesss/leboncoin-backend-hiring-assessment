package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterSwagger(t *testing.T) {
	spec := []byte("openapi: 3.0.3")
	mux := http.NewServeMux()
	RegisterSwagger(mux, spec)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("serves the Swagger UI page", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/swagger")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
		assert.Contains(t, string(body), "/swagger/openapi.yaml")
	})

	t.Run("serves the OpenAPI spec", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/swagger/openapi.yaml")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, spec, body)
	})
}
