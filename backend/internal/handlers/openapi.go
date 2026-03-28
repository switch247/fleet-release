package handlers

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed spec.yaml
var openAPISpec []byte

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>FleetLease API Docs</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        SwaggerUIBundle({
          url: '/docs/spec',
          dom_id: '#swagger',
          presets: [SwaggerUIBundle.presets.apis],
          layout: 'BaseLayout',
          deepLinking: true
        });
      };
    </script>
  </body>
</html>`

func (h *Handler) OpenAPIDocsPage(c echo.Context) error {
	return c.HTML(http.StatusOK, swaggerHTML)
}

func (h *Handler) OpenAPISpec(c echo.Context) error {
	return c.Blob(http.StatusOK, "text/yaml", openAPISpec)
}
