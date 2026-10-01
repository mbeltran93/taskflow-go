// Package openapispec embeds the API's OpenAPI document into the compiled
// binary (via go:embed) so the running server can serve it - and the Swagger
// UI built on top of it - without depending on any file being present next
// to the executable at runtime, which matters for a `scratch` Docker image.
package openapispec

import _ "embed"

//go:embed openapi.yaml
var Spec []byte
