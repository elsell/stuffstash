package evaluationrun

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func Verify(t *testing.T, repository ports.EvaluationRunRepository) {
	t.Helper()
	ctx := context.Background()
	run := Run(t, "primary")
	firstAudit := Record(t, "run-created", "primary", audit.ActionConversationEvaluationRunCreated)
	if err := repository.SaveEvaluationRun(ctx, run, 0, firstAudit); err != nil {
		t.Fatal(err)
	}
	saved, found, err := repository.EvaluationRun(ctx, TenantID, "primary")
	if err != nil || !found || !reflect.DeepEqual(saved.Snapshot(), run.Snapshot()) {
		t.Fatalf("rich round trip: %v found=%v", err, found)
	}
	if _, found, err := repository.EvaluationRun(ctx, "outside", "primary"); err != nil || found {
		t.Fatal("cross-tenant run exposed")
	}
	claimed, err := run.Claim("worker", Now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := Record(t, "run-created", "primary", audit.ActionConversationEvaluationRunProgressed)
	if err := repository.SaveEvaluationRun(ctx, claimed, 1, duplicate); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	unchanged, _, err := repository.EvaluationRun(ctx, TenantID, "primary")
	if err != nil || unchanged.Snapshot().Version != 1 {
		t.Fatal("failed audit changed run")
	}
	if err := repository.SaveEvaluationRun(ctx, claimed, 1, Record(t, "run-claimed", "primary", audit.ActionConversationEvaluationRunProgressed)); err != nil {
		t.Fatal(err)
	}
	if err := repository.SaveEvaluationRun(ctx, claimed, 1, Record(t, "stale-claim", "primary", audit.ActionConversationEvaluationRunProgressed)); !errors.Is(err, ports.ErrEvaluationRunConflict) {
		t.Fatalf("stale write: %v", err)
	}
	loaded, _, err := repository.EvaluationRun(ctx, TenantID, "primary")
	if err != nil {
		t.Fatal(err)
	}
	outcome := model.EvaluationObservedOutcome{Kind: model.EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []model.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "box"}}}
	partial, err := loaded.RecordCase("worker", 0, outcome, 2, time.Second, Now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.SaveEvaluationRun(ctx, partial, 2, Record(t, "case-completed", "primary", audit.ActionConversationEvaluationRunProgressed)); err != nil {
		t.Fatal(err)
	}
	cancelled, err := partial.Cancel(Now.Add(2 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.SaveEvaluationRun(ctx, cancelled, 3, Record(t, "run-cancelled", "primary", audit.ActionConversationEvaluationRunCancelled)); err != nil {
		t.Fatal(err)
	}
	late, err := partial.RecordCase("worker", 1, outcome, 1, time.Second, Now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.SaveEvaluationRun(ctx, late, 3, Record(t, "late-result", "primary", audit.ActionConversationEvaluationRunProgressed)); !errors.Is(err, ports.ErrEvaluationRunConflict) {
		t.Fatal("cancelled work overwritten")
	}
	// Independent validity must not permit resurrecting a terminal record.
	forged := claimed.Snapshot()
	forged.Version = 5
	forged.UpdatedAt = Now.Add(3 * time.Second)
	resurrection, err := model.RestoreEvaluationRun(forged)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.SaveEvaluationRun(ctx, resurrection, 4, Record(t, "resurrection", "primary", audit.ActionConversationEvaluationRunProgressed)); err == nil {
		t.Fatal("terminal run revived")
	}
	for _, id := range []string{"a-queued", "z-queued"} {
		if err := repository.SaveEvaluationRun(ctx, Run(t, id), 0, Record(t, id, id, audit.ActionConversationEvaluationRunCreated)); err != nil {
			t.Fatal(err)
		}
	}
	heads, err := repository.ListEvaluationRuns(ctx, TenantID, ports.EvaluationRunPageRequest{Limit: 2})
	if err != nil || len(heads) != 2 || heads[0].ID != "a-queued" || heads[1].ID != "primary" || heads[1].CompletedCases != 1 || heads[1].PassedCases != 1 {
		t.Fatalf("head list: %+v %v", heads, err)
	}
	tail, err := repository.ListEvaluationRuns(ctx, TenantID, ports.EvaluationRunPageRequest{AfterID: "primary", Limit: 2})
	if err != nil || len(tail) != 1 || tail[0].ID != "z-queued" {
		t.Fatalf("head cursor: %+v %v", tail, err)
	}
	ready, err := repository.RunnableEvaluationRuns(ctx, Now, 1)
	if err != nil || len(ready) != 1 || ready[0].ID != "a-queued" {
		t.Fatalf("bounded queue: %+v %v", ready, err)
	}
	for _, limit := range []int{0, 101} {
		if _, err := repository.ListEvaluationRuns(ctx, TenantID, ports.EvaluationRunPageRequest{Limit: limit}); err == nil {
			t.Fatal("invalid list limit accepted")
		}
		if _, err := repository.RunnableEvaluationRuns(ctx, Now, limit); err == nil {
			t.Fatal("invalid queue limit accepted")
		}
	}
}

func VerifyConcurrentClaims(t *testing.T, repository ports.EvaluationRunRepository) {
	t.Helper()
	ctx := context.Background()
	run := Run(t, "concurrent")
	if err := repository.SaveEvaluationRun(ctx, run, 0, Record(t, "concurrent-created", "concurrent", audit.ActionConversationEvaluationRunCreated)); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, token := range []string{"worker-a", "worker-b"} {
		claim, err := run.Claim(token, Now, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		record := Record(t, token, "concurrent", audit.ActionConversationEvaluationRunProgressed)
		go func() { <-start; results <- repository.SaveEvaluationRun(ctx, claim, 1, record) }()
	}
	close(start)
	wins, conflicts := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			wins++
		} else if errors.Is(err, ports.ErrEvaluationRunConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("claims won=%d conflicts=%d", wins, conflicts)
	}
	ready, err := repository.RunnableEvaluationRuns(ctx, Now, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, ref := range ready {
		if ref.ID == "concurrent" {
			t.Fatal("live lease discoverable")
		}
	}
	ready, err = repository.RunnableEvaluationRuns(ctx, Now.Add(time.Minute), 100)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ref := range ready {
		if ref.ID == "concurrent" {
			found = true
		}
	}
	if !found {
		t.Fatal("expired lease not recoverable")
	}
}
