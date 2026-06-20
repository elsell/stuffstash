package httpserver

import (
	"net/http"
	"testing"
	"time"
)

func TestResponsesIncludeSecurityHeaders(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequest(server, http.MethodGet, "/", "", nil)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "no-referrer",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
	}
	for header, expected := range expectedHeaders {
		if response.Header().Get(header) != expected {
			t.Fatalf("expected %s header %q, got %q", header, expected, response.Header().Get(header))
		}
	}
}

func TestServerOptionsConfigureTimeoutsAndBodyLimit(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		MaxJSONBodyBytes:  8,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      4 * time.Second,
		IdleTimeout:       5 * time.Second,
	})

	if server.ReadHeaderTimeout != 2*time.Second || server.ReadTimeout != 3*time.Second || server.WriteTimeout != 4*time.Second || server.IdleTimeout != 5*time.Second {
		t.Fatalf("unexpected server timeouts: %+v", server)
	}

	response := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:user-one", map[string]string{"name": "this body is too large"})
	if response.Code < http.StatusBadRequest {
		t.Fatalf("expected oversized JSON body to fail, got %d", response.Code)
	}
}
