// Package specs exposes the OpenAPI description of the API.
package specs

import _ "embed"

// OpenAPI is the content of openapi.yaml, embedded in the binary at build time.
//
//go:embed openapi.yaml
var OpenAPI []byte
