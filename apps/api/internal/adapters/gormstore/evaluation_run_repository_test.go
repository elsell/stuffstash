package gormstore

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

func TestEvaluationRunRepository(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, evaluationrun.TenantID, "Home")
	evaluationrun.Verify(t, store)
}
