package httpserver

import (
	"net/http"
	"slices"
	"strings"
)

const browserAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"

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
			w.Header().Set("Access-Control-Allow-Methods", browserAllowedMethods)
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
	if method != "" && !slices.Contains(strings.Split(browserAllowedMethods, ", "), method) {
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
