package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/config"
	"strings"
	"testing"
)

func TestApplicationRejectsInvalidWorkflowLimitsBeforeConstruction(t *testing.T) {
	const variable = "STUFF_STASH_WORKFLOW_MAX_MODEL_CALLS"
	t.Setenv(variable, "invalid-sensitive-value")
	_, err := buildApplication(context.Background(), config.Load(), nil, nil, nil, repositories{})
	if err == nil || !strings.Contains(err.Error(), variable) || strings.Contains(err.Error(), "invalid-sensitive-value") {
		t.Fatalf("expected safe workflow startup error, got %v", err)
	}
}
