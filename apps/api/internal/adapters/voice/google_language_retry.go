package voice

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const googleStructuredInferenceAttempts = 2

func retryableGoogleLanguageInferenceError(err error) bool {
	var httpErr googleProviderHTTPError
	if errors.As(err, &httpErr) {
		return httpErr.statusCode == http.StatusTooManyRequests || httpErr.statusCode == http.StatusInternalServerError || httpErr.statusCode == http.StatusBadGateway || httpErr.statusCode == http.StatusServiceUnavailable || httpErr.statusCode == http.StatusGatewayTimeout
	}
	return false
}

func sleepGoogleLanguageRetry(ctx context.Context, attempt int, err error) error {
	delay := time.Duration(attempt+1) * 250 * time.Millisecond
	var httpErr googleProviderHTTPError
	if errors.As(err, &httpErr) && httpErr.retryAfter > 0 {
		delay = httpErr.retryAfter
	}
	const maxDelay = 2 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
