package agentmodel

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

type evaluationWorkerClock struct{ now time.Time }

func (c *evaluationWorkerClock) Now() time.Time { return c.now }

type evaluationWorkerProviders struct {
	err   error
	calls int
}

func (p *evaluationWorkerProviders) ResolveEvaluationRunProviders(context.Context, tenant.ID, model.EvaluationRun) (ports.EvaluationExecutionProviders, error) {
	p.calls++
	return ports.EvaluationExecutionProviders{}, p.err
}

type evaluationWorkerExecutor struct {
	calls    int
	outcomes []model.EvaluationObservedOutcome
	before   func()
}

func (e *evaluationWorkerExecutor) Execute(ctx context.Context, input ports.ConversationEvaluationInput) (ports.ConversationEvaluationResult, error) {
	if e.before != nil {
		e.before()
	}
	if err := ctx.Err(); err != nil {
		return ports.ConversationEvaluationResult{}, err
	}
	if e.calls >= len(e.outcomes) {
		return ports.ConversationEvaluationResult{}, ports.ErrInvalidProviderInput
	}
	result := ports.ConversationEvaluationResult{Outcome: e.outcomes[e.calls], Coverage: ports.EvaluationCoverageText, ModelCalls: 1, Duration: time.Second}
	e.calls++
	return result, nil
}
func evaluationWorkerSetup(t *testing.T) (EvaluationWorkerDependencies, *memory.Store, *evaluationWorkerExecutor, *evaluationWorkerClock) {
	t.Helper()
	deps, input, store := evaluationCommandSetup(t)
	_, err := NewEvaluationRunCommandService(deps).Queue(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	executor := &evaluationWorkerExecutor{outcomes: []model.EvaluationObservedOutcome{
		{Kind: model.EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []model.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "box"}}},
		{Kind: model.EvaluationOutcomeProposal, ReferencedAssets: []string{"clothes"}, Locations: []model.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "box"}}, Proposals: []model.EvaluationProposal{{Operation: model.OperationCheckout, TargetID: "clothes", Details: "For Sam"}}},
	}}
	clock := &evaluationWorkerClock{now: evaluationCommandClock{}.Now()}
	return EvaluationWorkerDependencies{Runs: store, Authorizer: allowTenantConfigureAuthorizer{}, Providers: &evaluationWorkerProviders{}, Executor: executor, IDs: &workflowSequenceIDs{next: 100}, Clock: clock, LeaseGrace: time.Minute}, store, executor, clock
}
func workerReference(t *testing.T, store *memory.Store) ports.EvaluationRunReference {
	t.Helper()
	rows, err := store.ListEvaluationRuns(context.Background(), fixture.TenantID, ports.EvaluationRunPageRequest{Limit: 100})
	if err != nil || len(rows) != 1 {
		t.Fatal("missing queued run")
	}
	return rows[0].EvaluationRunReference
}
func TestEvaluationWorkerCompletesIsolatedSuite(t *testing.T) {
	deps, store, executor, _ := evaluationWorkerSetup(t)
	ref := workerReference(t, store)
	if err := NewEvaluationWorker(deps).Process(context.Background(), ref); err != nil {
		t.Fatal(err)
	}
	run, _, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if err != nil || run.Snapshot().State != model.EvaluationRunSucceeded || len(run.Snapshot().Results) != 2 || executor.calls != 2 {
		t.Fatalf("suite did not pass: state=%s results=%d calls=%d err=%v", run.Snapshot().State, len(run.Snapshot().Results), executor.calls, err)
	}
	if err := NewEvaluationWorker(deps).Process(context.Background(), ref); err != nil || executor.calls != 2 {
		t.Fatal("terminal run executed again")
	}
}
func TestEvaluationWorkerRejectsRevocationAndConfigurationDrift(t *testing.T) {
	for _, revoked := range []bool{true, false} {
		deps, store, executor, _ := evaluationWorkerSetup(t)
		providers := deps.Providers.(*evaluationWorkerProviders)
		want := model.EvaluationRunFailureConfigurationChanged
		if revoked {
			deps.Authorizer = denyTenantAuthorizer{}
			want = model.EvaluationRunFailureAccessRevoked
		} else {
			providers.err = ports.ErrEvaluationConfigurationChanged
		}
		ref := workerReference(t, store)
		if err := NewEvaluationWorker(deps).Process(context.Background(), ref); err != nil {
			t.Fatal(err)
		}
		run, _, _ := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
		if run.Snapshot().FailureCode != want || executor.calls != 0 {
			t.Fatal("unsafe run reached executor")
		}
		if revoked && providers.calls != 0 {
			t.Fatal("revoked author reached providers")
		}
	}
}
func TestEvaluationWorkerCannotOverwriteCancellation(t *testing.T) {
	deps, store, executor, clock := evaluationWorkerSetup(t)
	ref := workerReference(t, store)
	executor.before = func() {
		run, _, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
		if err != nil {
			t.Fatal(err)
		}
		cancelled, err := run.Cancel(clock.Now())
		if err != nil {
			t.Fatal(err)
		}
		record := fixture.Record(t, "user-cancel", string(ref.ID), "conversation_evaluation_run.cancelled")
		if err := store.SaveEvaluationRun(context.Background(), cancelled, run.Snapshot().Version, record); err != nil {
			t.Fatal(err)
		}
	}
	if err := NewEvaluationWorker(deps).Process(context.Background(), ref); err != nil {
		t.Fatal(err)
	}
	run, _, _ := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if run.Snapshot().State != model.EvaluationRunCancelled || len(run.Snapshot().Results) != 0 || executor.calls != 1 {
		t.Fatal("late result overwrote cancellation")
	}
}

func TestEvaluationWorkerResumesOnlyUnfinishedCases(t *testing.T) {
	deps, store, executor, clock := evaluationWorkerSetup(t)
	ref := workerReference(t, store)
	run, _, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := run.Claim("old-worker", clock.Now(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvaluationRun(context.Background(), claimed, run.Snapshot().Version, fixture.Record(t, "old-claim", string(ref.ID), "conversation_evaluation_run.progressed")); err != nil {
		t.Fatal(err)
	}
	partial, err := claimed.RecordCase("old-worker", 0, executor.outcomes[0], 1, time.Second, clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvaluationRun(context.Background(), partial, claimed.Snapshot().Version, fixture.Record(t, "old-result", string(ref.ID), "conversation_evaluation_run.progressed")); err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(2 * time.Second)
	executor.outcomes = executor.outcomes[1:]
	if err := NewEvaluationWorker(deps).Process(context.Background(), ref); err != nil {
		t.Fatal(err)
	}
	final, _, _ := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if final.Snapshot().State != model.EvaluationRunSucceeded || final.Snapshot().Attempts != 2 || executor.calls != 1 {
		t.Fatal("completed prefix replayed or recovery failed")
	}
}
func TestEvaluationWorkerExhaustedLeaseFailsWithoutExecuting(t *testing.T) {
	deps, store, executor, clock := evaluationWorkerSetup(t)
	ref := workerReference(t, store)
	run, _, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"lost-one", "lost-two"} {
		claimed, err := run.Claim(token, clock.Now(), time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.SaveEvaluationRun(context.Background(), claimed, run.Snapshot().Version, fixture.Record(t, token, string(ref.ID), "conversation_evaluation_run.progressed")); err != nil {
			t.Fatal(err)
		}
		run = claimed
		clock.now = clock.now.Add(2 * time.Second)
	}
	if err := NewEvaluationWorker(deps).Process(context.Background(), ref); err != nil {
		t.Fatal(err)
	}
	final, _, _ := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if final.Snapshot().FailureCode != model.EvaluationRunFailureWorkerLost || executor.calls != 0 {
		t.Fatal("exhausted run executed or remained live")
	}
}

func TestEvaluationWorkerRejectsPositiveDurationOverflowBeforeClaim(t *testing.T) {
	deps, store, _, _ := evaluationWorkerSetup(t)
	input := fixture.Run(t, "overflow").Snapshot().Input
	revision := input.Workflow.Snapshot()
	settings := revision.Definition.Settings()
	settings.Budget.ElapsedSeconds = 18446744074
	revision.Limits.Budget.ElapsedSeconds = settings.Budget.ElapsedSeconds
	input.Limits = revision.Limits
	definition, err := model.NewWorkflowDefinition(settings, revision.Limits)
	if err != nil {
		t.Fatal(err)
	}
	revision.Definition = definition
	input.Workflow, err = model.NewWorkflowRevision(revision)
	if err != nil {
		t.Fatal(err)
	}
	run, err := model.NewEvaluationRun(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvaluationRun(context.Background(), run, 0, fixture.Record(t, "overflow-created", "overflow", "conversation_evaluation_run.created")); err != nil {
		t.Fatal(err)
	}
	ref := ports.EvaluationRunReference{TenantID: fixture.TenantID, ID: "overflow"}
	if err := NewEvaluationWorker(deps).Process(context.Background(), ref); err == nil {
		t.Fatal("positive overflow accepted")
	}
	saved, _, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if err != nil || saved.Snapshot().Version != 1 || deps.Providers.(*evaluationWorkerProviders).calls != 0 {
		t.Fatal("overflow run claimed or providers accessed")
	}
}
