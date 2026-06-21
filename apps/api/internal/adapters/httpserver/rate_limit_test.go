package httpserver

import (
	"net/http"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRateLimitReturnsSafeEnvelopeAndRetryAfter(t *testing.T) {
	observer := &fakeObserver{}
	server := NewServerWithOptions(":0", newTestApp(observer, "unused-id"), Options{
		RateLimiter: NewTokenBucketRateLimiter(1, time.Minute, 1),
		Observer:    observer,
	})

	first := performRequestFrom(server, http.MethodGet, "/healthz", "", nil, "203.0.113.10:1234")
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request to pass, got %d with body %s", first.Code, first.Body.String())
	}

	second := performRequestFrom(server, http.MethodGet, "/healthz", "", nil, "203.0.113.10:1234")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusTooManyRequests, second.Code, second.Body.String())
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
	assertSafeError(t, second, "rate_limited", "Too many requests.")
	if !observer.hasEvent(ports.EventHTTPRateLimited) {
		t.Fatal("expected rate limited observability event")
	}
}

func TestRateLimitCannotBeBypassedByRotatingBearerTokens(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		RateLimiter: NewTokenBucketRateLimiter(1, time.Minute, 1),
	})

	first := performRequestFrom(server, http.MethodGet, "/healthz", "Bearer dev:first", nil, "203.0.113.11:1234")
	if first.Code != http.StatusOK {
		t.Fatalf("expected first token request to pass, got %d", first.Code)
	}
	second := performRequestFrom(server, http.MethodGet, "/healthz", "Bearer totally-different-bogus-token", nil, "203.0.113.11:1234")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected rotated token to still be limited, got %d with body %s", second.Code, second.Body.String())
	}
}

func TestRateLimitCanBeExplicitlyDisabled(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		RateLimitDisabled: true,
		RateLimiter:       NewTokenBucketRateLimiter(1, time.Minute, 1),
	})

	first := performRequestFrom(server, http.MethodGet, "/healthz", "", nil, "203.0.113.12:1234")
	second := performRequestFrom(server, http.MethodGet, "/healthz", "", nil, "203.0.113.12:1234")
	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("expected disabled limiter to allow both requests, got %d and %d", first.Code, second.Code)
	}
}
