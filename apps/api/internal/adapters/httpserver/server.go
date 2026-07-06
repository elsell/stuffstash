package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func init() {
	huma.NewError = shared.NewErrorEnvelope
}

type Options struct {
	CORSAllowedOrigins          []string
	MobileAuth                  MobileAuthOptions
	MaxJSONBodyBytes            int64
	RateLimitDisabled           bool
	RateLimiter                 ports.RateLimiter
	RateLimitRequests           int
	RateLimitWindow             time.Duration
	RateLimitBurst              int
	Observer                    ports.Observer
	ReadHeaderTimeout           time.Duration
	ReadTimeout                 time.Duration
	WriteTimeout                time.Duration
	IdleTimeout                 time.Duration
	RealtimeVoiceSessionTimeout time.Duration
}

func NewServer(addr string, application app.App) *http.Server {
	return NewServerWithOptions(addr, application, Options{})
}

func NewServerWithOptions(addr string, application app.App, options Options) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /healthz", handleHealth(application))
	mux.HandleFunc("GET /.well-known/stuff-stash/mobile-auth", handleMobileAuthMetadata(options.MobileAuth))
	mux.HandleFunc("GET "+realtimeVoicePath, handleRealtimeVoice(application, normalizeDuration(options.RealtimeVoiceSessionTimeout, 60*time.Second)))

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

	maxJSONBodyBytes := options.MaxJSONBodyBytes
	if maxJSONBodyBytes <= 0 {
		maxJSONBodyBytes = 1024 * 1024
	}
	applyJSONBodyLimit(api, maxJSONBodyBytes)
	rateLimiter := options.RateLimiter
	if options.RateLimitDisabled {
		rateLimiter = nil
	} else if rateLimiter == nil {
		rateLimiter = NewTokenBucketRateLimiter(options.RateLimitRequests, options.RateLimitWindow, options.RateLimitBurst)
	}
	handler := withSecurityHeaders(withCORS(withRateLimit(mux, rateLimiter, options.Observer), options.CORSAllowedOrigins))
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: normalizeDuration(options.ReadHeaderTimeout, 5*time.Second),
		ReadTimeout:       normalizeDuration(options.ReadTimeout, 15*time.Second),
		WriteTimeout:      normalizeDuration(options.WriteTimeout, 30*time.Second),
		IdleTimeout:       normalizeDuration(options.IdleTimeout, 60*time.Second),
	}
}

func applyJSONBodyLimit(api huma.API, maxBytes int64) {
	if maxBytes <= 0 {
		maxBytes = 1024 * 1024
	}
	humaLimit := maxBytes + 1
	for routePath, path := range api.OpenAPI().Paths {
		for _, operation := range []*huma.Operation{
			path.Get,
			path.Put,
			path.Post,
			path.Delete,
			path.Options,
			path.Head,
			path.Patch,
			path.Trace,
		} {
			if operation != nil && !isAttachmentCreateRoute(routePath, operation) {
				operation.MaxBodyBytes = humaLimit
			}
		}
	}
}

func isAttachmentCreateRoute(routePath string, operation *huma.Operation) bool {
	return operation.Method == http.MethodPost && strings.HasSuffix(routePath, "/attachments")
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func normalizeDuration(value time.Duration, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	allowed := map[string]struct{}{}
	for _, origin := range allowedOrigins {
		if origin == "" {
			continue
		}
		allowed[origin] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			if origin != "" && !validCORSPreflight(r) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validCORSPreflight(r *http.Request) bool {
	method := strings.ToUpper(strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")))
	switch method {
	case "":
		return true
	case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete:
	default:
		return false
	}

	for _, requestedHeader := range strings.Split(r.Header.Get("Access-Control-Request-Headers"), ",") {
		requestedHeader = http.CanonicalHeaderKey(strings.TrimSpace(requestedHeader))
		if requestedHeader == "" {
			continue
		}
		switch requestedHeader {
		case "Authorization", "Content-Type", "X-Request-Id":
		default:
			return false
		}
	}
	return true
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
