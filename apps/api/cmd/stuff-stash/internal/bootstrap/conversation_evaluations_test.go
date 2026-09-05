package bootstrap

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/config"
)

func TestBuildApplicationRejectsInvalidEvaluationConfiguration(t *testing.T) {
	t.Setenv("STUFF_STASH_EVALUATION_CONCURRENCY", "9")
	_, err := buildApplication(context.Background(), config.Load(), nil, nil, nil, repositories{})
	if err == nil || !strings.Contains(err.Error(), "STUFF_STASH_EVALUATION_CONCURRENCY") {
		t.Fatalf("invalid worker config not rejected first: %v", err)
	}
}
