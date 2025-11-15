package swagger

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	apicontract "github.com/tuanvumaihuynh/outbox-pattern/api-contract"
)

const (
	// swaggerURL is the URL path where the Swagger UI will be served
	swaggerURL = "/docs"

	// swaggerSpecURL is the URL path where the OpenAPI specification will be served
	swaggerSpecURL = "/docs/openapi.yml"
)

// Register registers the swagger handler on the given router
func Register(r chi.Router) {
	template := getTemplate(swaggerSpecURL)
	templateBytes := []byte(template)

	r.Get(swaggerURL, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write(templateBytes)
	})

	specBytes := apicontract.GetSpecBytes()
	r.Get(swaggerSpecURL, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write(specBytes)
	})
}

// getTemplate returns the HTML template for Swagger UI
func getTemplate(specPath string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="SwaggerUI" />
  <title>SwaggerUI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.29.3/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://static1.smartbear.co/swagger/media/assets/swagger_fav.png" sizes="32x32" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.29.3/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '%s',
      dom_id: '#swagger-ui',
      deepLinking: true,
	  showExtensions: true,
	  showCommonExtensions: true,
    });
  };
</script>
</body>
</html>
`, specPath)
}
