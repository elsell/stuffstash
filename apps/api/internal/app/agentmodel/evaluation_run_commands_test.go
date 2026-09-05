package agentmodel

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

type evaluationCommandClock struct{}

func (evaluationCommandClock) Now() time.Time { return fixture.Now.Add(time.Minute) }

type evaluationProviderSnapshots struct {
	values []domain.EvaluationRunProvider
	err    error
}

func (f evaluationProviderSnapshots) SnapshotEvaluationProviders(_ context.Context, id tenant.ID, revision domain.WorkflowRevision) ([]domain.EvaluationRunProvider, error) {
	if revision.Snapshot().TenantID != domain.TenantID(id) {
		return nil, ports.ErrForbidden
	}
	return append([]domain.EvaluationRunProvider(nil), f.values...), f.err
}

func evaluationCommandSetup(t *testing.T) (EvaluationRunCommandDependencies, QueueEvaluationRunInput, *memory.Store) {
	t.Helper()
	ctx := context.Background()
	store := memory.NewStore()
	name, _ := tenant.NewName("Home")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: fixture.TenantID, Name: name}); err != nil {
		t.Fatal(err)
	}
	pinned := fixture.Run(t, "template").Snapshot().Input
	record := fixture.Record(t, "workflow-audit", "workflow", audit.ActionConversationWorkflowRevisionCreated)
	record.TargetType = audit.TargetConversationWorkflow
	if err := store.AppendWorkflowRevision(ctx, pinned.Workflow, 0, record); err != nil {
		t.Fatal(err)
	}
	input := QueueEvaluationRunInput{EvaluationRunAccess: EvaluationRunAccess{Principal: testPrincipal(), TenantID: fixture.TenantID, Source: audit.SourceAPI}, WorkflowID: pinned.Workflow.Snapshot().WorkflowID, RevisionID: pinned.Workflow.Snapshot().ID}
	for _, revision := range pinned.Cases {
		value := revision.Snapshot()
		record := fixture.Record(t, "case-"+string(value.ID), string(value.CaseID), audit.ActionConversationEvaluationCaseRevisionCreated)
		record.TargetType = audit.TargetConversationEvaluationCase
		if err := store.AppendEvaluationCaseRevision(ctx, revision, 0, record); err != nil {
			t.Fatal(err)
		}
		input.Cases = append(input.Cases, EvaluationRunCaseReference{CaseID: value.CaseID, RevisionID: value.ID})
	}
	deps := EvaluationRunCommandDependencies{Authorizer: allowTenantConfigureAuthorizer{}, Runs: store, Workflows: store, Cases: store, Providers: evaluationProviderSnapshots{values: pinned.Providers}, IDs: &workflowSequenceIDs{}, Clock: evaluationCommandClock{}, Limits: pinned.Limits, MaxAttempts: 2}
	return deps, input, store
}

func TestEvaluationRunCommandsPinRevisionsAndCancelWithVersion(t *testing.T) {
	deps, input, store := evaluationCommandSetup(t)
	service := NewEvaluationRunCommandService(deps)
	ctx := context.Background()
	run, err := service.Queue(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	data := run.Snapshot()
	expected := fixture.Run(t, "template").Snapshot().Input
	if data.State != domain.EvaluationRunQueued || data.Input.AuthorID != domain.WorkflowAuthorID(input.Principal.ID) || !reflect.DeepEqual(data.Input.Cases, expected.Cases) || !reflect.DeepEqual(data.Input.Workflow, expected.Workflow) || !reflect.DeepEqual(data.Input.Providers, expected.Providers) {
		t.Fatal("queued run did not pin configuration")
	}
	cancellation := CancelEvaluationRunInput{EvaluationRunAccess: input.EvaluationRunAccess, RunID: data.Input.ID, ExpectedVersion: 1}
	result, err := service.Cancel(ctx, cancellation)
	if err != nil || result.Snapshot().State != domain.EvaluationRunCancelled {
		t.Fatalf("cancel: %v", err)
	}
	if _, err := service.Cancel(ctx, cancellation); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("stale cancel: %v", err)
	}
	cancellation.ExpectedVersion = 2
	if _, err := service.Cancel(ctx, cancellation); err != nil {
		t.Fatalf("idempotent cancel: %v", err)
	}
	records, err := store.ListTenantAuditRecords(ctx, fixture.TenantID, ports.AuditRecordPageRequest{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	created, cancelled := 0, 0
	for _, record := range records {
		if record.Action == audit.ActionConversationEvaluationRunCreated {
			created++
		}
		if record.Action == audit.ActionConversationEvaluationRunCancelled {
			cancelled++
		}
	}
	if created != 1 || cancelled != 1 {
		t.Fatalf("command audit counts %d/%d", created, cancelled)
	}
}

func TestEvaluationRunQueueRejectsInvalidPinsAndProviderFailure(t *testing.T) {
	for _, scenario := range []string{"missing workflow", "missing case", "wrong tenant", "duplicate case", "empty cases", "provider unavailable", "invalid provider snapshot"} {
		t.Run(scenario, func(t *testing.T) {
			deps, input, store := evaluationCommandSetup(t)
			switch scenario {
			case "missing workflow":
				input.RevisionID = "missing"
			case "missing case":
				input.Cases[0].RevisionID = "missing"
			case "wrong tenant":
				input.TenantID = "outside"
			case "duplicate case":
				input.Cases = append(input.Cases, input.Cases[0])
			case "empty cases":
				input.Cases = nil
			case "provider unavailable":
				deps.Providers = evaluationProviderSnapshots{err: ports.ErrInvalidProviderInput}
			case "invalid provider snapshot":
				deps.Providers = evaluationProviderSnapshots{}
			}
			if _, err := NewEvaluationRunCommandService(deps).Queue(context.Background(), input); err == nil {
				t.Fatal("invalid queue accepted")
			}
			rows, err := store.ListEvaluationRuns(context.Background(), fixture.TenantID, ports.EvaluationRunPageRequest{Limit: 100})
			if err != nil || len(rows) != 0 {
				t.Fatal("failed queue persisted a run")
			}
		})
	}
}

func TestEvaluationRunCommandsAuthorizeBeforeDependencies(t *testing.T) {
	for _, anonymous := range []bool{false, true} {
		access := EvaluationRunAccess{Principal: testPrincipal(), TenantID: fixture.TenantID, Source: audit.SourceAPI}
		want := ports.ErrForbidden
		if anonymous {
			access.Principal = identity.Principal{}
			want = apperrors.ErrUnauthenticated
		}
		service := NewEvaluationRunCommandService(EvaluationRunCommandDependencies{Authorizer: denyTenantAuthorizer{}})
		if _, err := service.Queue(context.Background(), QueueEvaluationRunInput{EvaluationRunAccess: access}); !errors.Is(err, want) {
			t.Fatalf("unauthorized queue: %v", err)
		}
		if _, err := service.Cancel(context.Background(), CancelEvaluationRunInput{EvaluationRunAccess: access}); !errors.Is(err, want) {
			t.Fatalf("unauthorized cancel: %v", err)
		}
	}
}

// The read view can lag a worker's commit; the real repository still fences writes.
type evaluationStaleRunView struct {
	ports.EvaluationRunRepository
	previous domain.EvaluationRun
}

func (v evaluationStaleRunView) EvaluationRun(context.Context, tenant.ID, domain.EvaluationRunID) (domain.EvaluationRun, bool, error) {
	return v.previous, true, nil
}

type evaluationDuplicateAuditIDs struct{}

func (evaluationDuplicateAuditIDs) NewID() string { return "workflow-audit" }

func TestEvaluationRunCommandsPreserveAtomicAuditAndConcurrentProgress(t *testing.T) {
	deps, input, store := evaluationCommandSetup(t)
	ctx := context.Background()
	failed := deps
	failed.IDs = evaluationDuplicateAuditIDs{}
	if _, err := NewEvaluationRunCommandService(failed).Queue(ctx, input); err == nil {
		t.Fatal("duplicate audit accepted")
	}
	rows, err := store.ListEvaluationRuns(ctx, input.TenantID, ports.EvaluationRunPageRequest{Limit: 100})
	if err != nil || len(rows) != 0 {
		t.Fatal("audit failure persisted a run")
	}
	queued, err := NewEvaluationRunCommandService(deps).Queue(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := queued.Claim("worker", evaluationCommandClock{}.Now(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvaluationRun(ctx, claimed, 1, fixture.Record(t, "worker-claim", string(queued.Snapshot().Input.ID), audit.ActionConversationEvaluationRunProgressed)); err != nil {
		t.Fatal(err)
	}
	deps.Runs = evaluationStaleRunView{EvaluationRunRepository: store, previous: queued}
	_, err = NewEvaluationRunCommandService(deps).Cancel(ctx, CancelEvaluationRunInput{EvaluationRunAccess: input.EvaluationRunAccess, RunID: queued.Snapshot().Input.ID, ExpectedVersion: 1})
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("concurrent claim not fenced: %v", err)
	}
	stored, found, err := store.EvaluationRun(ctx, input.TenantID, queued.Snapshot().Input.ID)
	if err != nil || !found || stored.Snapshot().State != domain.EvaluationRunRunning {
		t.Fatal("cancellation overwrote worker progress")
	}
}
