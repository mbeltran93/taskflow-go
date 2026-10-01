package handler

import (
	"fmt"
	"net/http"

	"github.com/mbeltran93/taskflow-go/internal/openapispec"
)

// ServeOpenAPISpec handles GET /openapi.yaml.
func ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(openapispec.Spec)
}

// ServeSwaggerUI handles GET /swagger: a static HTML page that loads
// swagger-ui-dist from a CDN and points it at /openapi.yaml. No code
// generation step is needed to keep this in sync - it just reads whatever
// the server itself exposes.
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, swaggerUIPage)
}

const swaggerUIPage = `<!DOCTYPE html>
<html>
<head>
  <title>TaskFlow API docs</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/openapi.yaml',
        dom_id: '#swagger-ui',
      });
    };
  </script>
</body>
</html>`
