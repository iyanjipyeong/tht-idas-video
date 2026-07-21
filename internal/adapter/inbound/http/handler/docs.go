package handler

import (
	"net/http"
	"os"

	"idas-video/internal/entity"
)

type OpenAPIHandler struct{}

func NewOpenAPIHandler() *OpenAPIHandler { return &OpenAPIHandler{} }

func (handler *OpenAPIHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	openAPISpec, err := os.ReadFile("docs/openapi.yaml")
	if err != nil {
		writeError(writer, err)
		return
	}

	writer.Header().Set(entity.HTTPHeaderContentType, entity.HTTPContentTypeYAMLUTF8)
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(openAPISpec)
}

type SwaggerUIHandler struct{}

func NewSwaggerUIHandler() *SwaggerUIHandler { return &SwaggerUIHandler{} }

func (handler *SwaggerUIHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set(entity.HTTPHeaderContentType, entity.HTTPContentTypeHTMLUTF8)
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(swaggerHTML))
}

const swaggerHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>IDAS Video Subscription API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function () {
        window.ui = SwaggerUIBundle({
          url: '/openapi.yaml',
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [SwaggerUIBundle.presets.apis],
          layout: 'BaseLayout'
        });
      };
    </script>
  </body>
</html>`
