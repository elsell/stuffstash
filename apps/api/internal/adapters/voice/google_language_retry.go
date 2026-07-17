package voice

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func googleLanguageInferenceAttempts(input ports.LanguageInferenceInput) int {
	if input.PlanOnly || input.FinalOnly || input.Investigation != nil {
		return 2
	}
	return 1
}

func retryableGoogleLanguageInferenceError(err error) bool {
	var httpErr googleProviderHTTPError
	if errors.As(err, &httpErr) {
		return httpErr.statusCode == http.StatusTooManyRequests || httpErr.statusCode == http.StatusInternalServerError || httpErr.statusCode == http.StatusBadGateway || httpErr.statusCode == http.StatusServiceUnavailable || httpErr.statusCode == http.StatusGatewayTimeout
	}
	return false
}

func retryableGoogleStructuredOutputError(input ports.LanguageInferenceInput, err error) bool {
	if err == nil {
		return false
	}
	return input.PlanOnly || input.FinalOnly || input.Investigation != nil
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
