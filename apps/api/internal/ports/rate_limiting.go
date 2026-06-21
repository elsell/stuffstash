package ports

import (
	"context"
	"time"
)

type RateLimitDecision struct {
	Allowed    bool
	RetryAfter time.Duration
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, now time.Time) RateLimitDecision
}
