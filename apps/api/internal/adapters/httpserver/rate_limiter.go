package httpserver

import (
	"context"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type TokenBucketRateLimiter struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	burst       int
	buckets     map[string]rateLimitBucket
	lastCleanup time.Time
}

type rateLimitBucket struct {
	tokens    float64
	updatedAt time.Time
}

func NewTokenBucketRateLimiter(limit int, window time.Duration, burst int) *TokenBucketRateLimiter {
	defaultedLimit := false
	if limit <= 0 {
		limit = config.DefaultHTTPRateLimitRequests
		defaultedLimit = true
	}
	if window <= 0 {
		window = time.Minute
	}
	if burst <= 0 {
		if defaultedLimit {
			burst = config.DefaultHTTPRateLimitBurst
		} else {
			burst = limit
		}
	}
	return &TokenBucketRateLimiter{
		limit:   limit,
		window:  window,
		burst:   burst,
		buckets: map[string]rateLimitBucket{},
	}
}

func (l *TokenBucketRateLimiter) Allow(_ context.Context, key string, now time.Time) ports.RateLimitDecision {
	if strings.TrimSpace(key) == "" {
		key = "anonymous"
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanupExpiredBuckets(now)

	bucket := l.buckets[key]
	if bucket.updatedAt.IsZero() {
		bucket.tokens = float64(l.burst)
		bucket.updatedAt = now
	}

	elapsed := now.Sub(bucket.updatedAt)
	if elapsed > 0 {
		refillRate := float64(l.limit) / l.window.Seconds()
		bucket.tokens = math.Min(float64(l.burst), bucket.tokens+elapsed.Seconds()*refillRate)
		bucket.updatedAt = now
	}

	if bucket.tokens >= 1 {
		bucket.tokens--
		l.buckets[key] = bucket
		return ports.RateLimitDecision{Allowed: true}
	}

	l.buckets[key] = bucket
	refillRate := float64(l.limit) / l.window.Seconds()
	retryAfter := time.Duration(math.Ceil((1 - bucket.tokens) / refillRate * float64(time.Second)))
	if retryAfter <= 0 {
		retryAfter = time.Second
	}
	return ports.RateLimitDecision{Allowed: false, RetryAfter: retryAfter}
}

func (l *TokenBucketRateLimiter) cleanupExpiredBuckets(now time.Time) {
	if !l.lastCleanup.IsZero() && now.Sub(l.lastCleanup) < l.window {
		return
	}
	l.lastCleanup = now
	maxIdle := l.window * 2
	for key, bucket := range l.buckets {
		if bucket.updatedAt.IsZero() || now.Sub(bucket.updatedAt) <= maxIdle {
			continue
		}
		delete(l.buckets, key)
	}
}

func withRateLimit(next http.Handler, limiter ports.RateLimiter, observer ports.Observer) http.Handler {
	if limiter == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keyKind, key := rateLimitKey(r)
		decision := limiter.Allow(r.Context(), key, time.Now().UTC())
		if decision.Allowed {
			next.ServeHTTP(w, r)
			return
		}

		retryAfterSeconds := retryAfterSeconds(decision.RetryAfter)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(shared.ErrorEnvelope{
			BodyError: shared.ErrorBody{
				Code:    "rate_limited",
				Message: "Too many requests.",
				Details: []shared.ErrorDetail{},
			},
			Meta: shared.Meta{},
		})

		if observer != nil {
			observer.Record(r.Context(), ports.Event{
				Name:    ports.EventHTTPRateLimited,
				Message: "HTTP request rate limited",
				Fields: map[string]string{
					"key_kind":            keyKind,
					"retry_after_seconds": strconv.Itoa(retryAfterSeconds),
				},
			})
		}
	})
}

func retryAfterSeconds(duration time.Duration) int {
	if duration <= 0 {
		return 1
	}
	seconds := int(math.Ceil(duration.Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}

func rateLimitKey(r *http.Request) (string, string) {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil || strings.TrimSpace(host) == "" {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	if host == "" {
		host = "unknown"
	}
	return "remote_addr", "ip:" + host
}
