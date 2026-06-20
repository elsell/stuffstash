package httpserver

import (
	"net/http"
	"testing"
)

func TestCORSIsDenyByDefault(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequestWithHeaders(server, http.MethodGet, "/healthz", "", map[string]string{
		"Origin": "http://localhost:5173",
	}, nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no CORS allow origin by default, got %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		CORSAllowedOrigins: []string{"http://localhost:5173"},
	})

	response := performRequestWithHeaders(server, http.MethodOptions, "/tenants", "", map[string]string{
		"Origin":                        "http://localhost:5173",
		"Access-Control-Request-Method": "POST",
	}, nil)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected preflight status %d, got %d with body %s", http.StatusNoContent, response.Code, response.Body.String())
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("expected configured allow origin, got %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
	if response.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("expected CORS credentials to stay disabled, got %q", response.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestCORSDeniesUnconfiguredOrigin(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		CORSAllowedOrigins: []string{"http://localhost:5173"},
	})

	response := performRequestWithHeaders(server, http.MethodOptions, "/tenants", "", map[string]string{
		"Origin":                        "http://evil.localhost",
		"Access-Control-Request-Method": "POST",
	}, nil)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected preflight status %d, got %d with body %s", http.StatusNoContent, response.Code, response.Body.String())
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no CORS allow origin for unconfigured origin, got %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSRejectsUnsupportedPreflightMethod(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		CORSAllowedOrigins: []string{"http://localhost:5173"},
	})

	response := performRequestWithHeaders(server, http.MethodOptions, "/tenants", "", map[string]string{
		"Origin":                        "http://localhost:5173",
		"Access-Control-Request-Method": "PUT",
	}, nil)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected preflight status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
	}
}

func TestCORSRejectsUnsupportedPreflightHeaders(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		CORSAllowedOrigins: []string{"http://localhost:5173"},
	})

	response := performRequestWithHeaders(server, http.MethodOptions, "/tenants", "", map[string]string{
		"Origin":                         "http://localhost:5173",
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "Authorization, X-Forbidden",
	}, nil)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected preflight status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
	}
}
