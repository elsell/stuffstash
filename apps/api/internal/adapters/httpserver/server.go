package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
)

func init() {
	huma.NewError = shared.NewErrorEnvelope
}

func NewServer(addr string, application app.App) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /healthz", handleHealth(application))

	config := huma.DefaultConfig("Stuff Stash API", "0.1.0")
	config.DocsPath = "/docs"
	config.OpenAPIPath = "/openapi"
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "dev",
		},
	}

	api := humago.New(mux, config)
	registerRoutes(api, application)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(shared.SuccessEnvelope[indexResponse]{
		Data: indexResponse{
			Service: "stuff-stash",
			Links: indexLinksResponse{
				Health:  "/healthz",
				OpenAPI: "/openapi.json",
				Docs:    "/docs",
			},
		},
		Meta: shared.Meta{},
	})
}

func handleHealth(application app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := application.Health(r.Context())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service":"` + string(status.Service) + `","status":"` + string(status.Status) + `"}` + "\n"))
	}
}

type indexResponse struct {
	Service string             `json:"service"`
	Links   indexLinksResponse `json:"links"`
}

type indexLinksResponse struct {
	Health  string `json:"health"`
	OpenAPI string `json:"openapi"`
	Docs    string `json:"docs"`
}
