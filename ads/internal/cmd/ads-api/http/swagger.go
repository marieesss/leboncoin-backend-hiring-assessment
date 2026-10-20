package http

import (
	"net/http"

	httpswagger "github.com/swaggo/http-swagger/v2"
)

// RegisterSwagger serves Swagger UI on GET /swagger/ and the OpenAPI spec it displays on GET /swagger/openapi.yaml.
func RegisterSwagger(mux *http.ServeMux, spec []byte) {
	mux.Handle("GET /swagger/", httpswagger.Handler(httpswagger.URL("/swagger/openapi.yaml")))
	mux.HandleFunc("GET /swagger/openapi.yaml", func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write(spec)
	})
}
