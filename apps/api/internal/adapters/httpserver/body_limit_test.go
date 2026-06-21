package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOversizedJSONBodyReturnsSafeEnvelope(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{MaxJSONBodyBytes: 20})

	request := httptest.NewRequest(http.MethodPost, "/tenants", strings.NewReader(`{"name":"this body is too large"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer dev:owner")

	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusRequestEntityTooLarge, response.Code, response.Body.String())
	}
	assertSafeError(t, response, "payload_too_large", "Request body too large.")
}

func TestJSONBodyLimitAllowsExactConfiguredByteLength(t *testing.T) {
	const body = `{"name":"exact"}`
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "tenant-exact-id"), Options{MaxJSONBodyBytes: int64(len(body))})

	request := httptest.NewRequest(http.MethodPost, "/tenants", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer dev:owner")

	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected exact-limit body status %d, got %d with body %s", http.StatusCreated, response.Code, response.Body.String())
	}
}
