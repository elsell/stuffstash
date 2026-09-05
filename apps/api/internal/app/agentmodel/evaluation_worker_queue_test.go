package agentmodel

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

type evaluationQueueProviders struct{}

func (evaluationQueueProviders) ResolveEvaluationRunProviders(context.Context, tenant.ID, model.EvaluationRun) (ports.EvaluationExecutionProviders, error) {
	return ports.EvaluationExecutionProviders{}, nil
}

type evaluationConcurrentExecutor struct {
	active  atomic.Int32
	maximum atomic.Int32
	started chan struct{}
	release chan struct{}
}

func (e *evaluationConcurrentExecutor) Execute(ctx context.Context, input ports.ConversationEvaluationInput) (ports.ConversationEvaluationResult, error) {
	active := e.active.Add(1)
	defer e.active.Add(-1)
	for {
		previous := e.maximum.Load()
		if active <= previous || e.maximum.CompareAndSwap(previous, active) {
			break
		}
	}
	e.started <- struct{}{}
	select {
	case <-ctx.Done():
		return ports.ConversationEvaluationResult{}, ctx.Err()
	case <-e.release:
	}
	outcome := model.EvaluationObservedOutcome{Kind: model.EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []model.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "box"}}}
	if input.Case.Settings().Utterance == "Lend the baby clothes to Sam" {
		outcome.Kind = model.EvaluationOutcomeProposal
		outcome.Proposals = []model.EvaluationProposal{{Operation: model.OperationCheckout, TargetID: "clothes", Details: "For Sam"}}
	}
	return ports.ConversationEvaluationResult{Outcome: outcome, Coverage: ports.EvaluationCoverageText, ModelCalls: 1}, nil
}

type evaluationAtomicIDs struct{ next atomic.Int64 }

func (ids *evaluationAtomicIDs) NewID() string {
	return "queue-" + strconv.FormatInt(ids.next.Add(1), 10)
}
func TestEvaluationQueueBoundsConcurrencyAndCompletesDiscoveredRuns(t *testing.T) {
	deps, store, _, _ := evaluationWorkerSetup(t)
	for _, id := range []string{"second", "third"} {
		if err := store.SaveEvaluationRun(context.Background(), fixture.Run(t, id), 0, fixture.Record(t, "created-"+id, id, audit.ActionConversationEvaluationRunCreated)); err != nil {
			t.Fatal(err)
		}
	}
	executor := &evaluationConcurrentExecutor{started: make(chan struct{}, 6), release: make(chan struct{})}
	deps.Executor = executor
	deps.Providers = evaluationQueueProviders{}
	deps.IDs = &evaluationAtomicIDs{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	finished := make(chan error, 1)
	go func() { finished <- NewEvaluationWorker(deps).Drain(ctx, 3, 2) }()
	for count := 0; count < 2; count++ {
		select {
		case <-executor.started:
		case <-ctx.Done():
			t.Fatal("bounded workers did not start")
		}
	}
	if executor.maximum.Load() != 2 {
		t.Fatal("two independent runs did not run concurrently")
	}
	close(executor.release)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	rows, err := store.ListEvaluationRuns(context.Background(), fixture.TenantID, ports.EvaluationRunPageRequest{Limit: 100})
	if err != nil || len(rows) != 3 {
		t.Fatal("missing runs")
	}
	for _, row := range rows {
		if row.State != model.EvaluationRunSucceeded {
			t.Fatalf("run %s incomplete", row.ID)
		}
	}
	if executor.maximum.Load() > 2 || executor.active.Load() != 0 {
		t.Fatal("queue exceeded concurrency or returned before work stopped")
	}
}
func TestEvaluationQueueRejectsUnboundedRequestsBeforeClaims(t *testing.T) {
	deps, store, _, _ := evaluationWorkerSetup(t)
	for _, input := range [][2]int{{0, 1}, {101, 1}, {1, 0}, {1, 9}} {
		if err := NewEvaluationWorker(deps).Drain(context.Background(), input[0], input[1]); err == nil {
			t.Fatal("unbounded queue accepted")
		}
	}
	ref := workerReference(t, store)
	run, _, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if err != nil || run.Snapshot().Version != 1 {
		t.Fatal("invalid drain claimed work")
	}
}
