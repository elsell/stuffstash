package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"

	"github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

func TestEvaluationRunRepository(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, evaluationrun.TenantID, "Home")
	evaluationrun.Verify(t, store)
	evaluationrun.VerifyQueueTimestampPrecision(t, store)
}

func TestEvaluationRunTimestampPrecision(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, evaluationrun.TenantID, "Home")
	verifyEvaluationRunTimestampPrecision(t, store)
}
func verifyEvaluationRunTimestampPrecision(t *testing.T, store Store) {
	t.Helper()
	ctx := context.Background()
	input := evaluationrun.Run(t, "precision").Snapshot().Input
	input.CreatedAt = input.CreatedAt.Add(123 * time.Nanosecond)
	run, err := agentmodel.NewEvaluationRun(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvaluationRun(ctx, run, 0, evaluationrun.Record(t, "precision", "precision", audit.ActionConversationEvaluationRunCreated)); err != nil {
		t.Fatal(err)
	}
	restored, found, err := store.EvaluationRun(ctx, evaluationrun.TenantID, "precision")
	if err != nil || !found || !restored.Snapshot().Input.CreatedAt.Equal(input.CreatedAt) {
		t.Fatalf("nanosecond round trip: %v", err)
	}
}
