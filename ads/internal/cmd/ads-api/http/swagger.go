package http

import (
	"net/http"
)

// swaggerPage loads Swagger UI from a CDN and points it to the spec served by this API.
const swaggerPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Ads API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: "/swagger/openapi.yaml", dom_id: "#swagger-ui" });
  </script>
</body>
</html>`

// RegisterSwagger serves Swagger UI on GET /swagger and the raw spec on GET /swagger/openapi.yaml.
func RegisterSwagger(mux *http.ServeMux, spec []byte) {
	mux.HandleFunc("GET /swagger", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = rw.Write([]byte(swaggerPage))
	})

	mux.HandleFunc("GET /swagger/openapi.yaml", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/yaml")
		_, _ = rw.Write(spec)
	})
}
