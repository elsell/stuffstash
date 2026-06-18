package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/stuffstash/stuff-stash/internal/app"
)

func NewServer(addr string, application app.App) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth(application))

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func handleHealth(application app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := application.Health(r.Context())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(healthResponse{
			Service: string(status.Service),
			Status:  string(status.Status),
		})
	}
}

type healthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}
