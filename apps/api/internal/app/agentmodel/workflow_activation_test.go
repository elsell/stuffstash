package agentmodel

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
	"testing"
	"time"
)

func TestWorkflowActivationAuthorizesBeforeDependencies(t *testing.T) {
	service := NewWorkflowActivationService(WorkflowActivationDependencies{Authorizer: denyTenantAuthorizer{}})
	_, err := service.Activate(context.Background(), ActivateWorkflowInput{EvaluationRunAccess: EvaluationRunAccess{Principal: testPrincipal(), TenantID: "home"}})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("unauthorized: %v", err)
	}
	_, err = service.Activate(context.Background(), ActivateWorkflowInput{})
	if !errors.Is(err, apperrors.ErrUnauthenticated) {
		t.Fatalf("anonymous: %v", err)
	}
}
func TestWorkflowActivationPersistsAuditedSelectionAndRejectsStaleEvidence(t *testing.T) {
	deps, queue, store := evaluationCommandSetup(t)
	ctx := context.Background()
	run, err := NewEvaluationRunCommandService(deps).Queue(ctx, queue)
	if err != nil {
		t.Fatal(err)
	}
	input := ActivateWorkflowInput{EvaluationRunAccess: queue.EvaluationRunAccess, WorkflowID: queue.WorkflowID, RevisionID: queue.RevisionID, RunID: run.Snapshot().Input.ID, Cases: queue.Cases}
	activation := WorkflowActivationDependencies{Authorizer: deps.Authorizer, Workflows: store, Runs: store, Providers: deps.Providers, IDs: deps.IDs, Clock: deps.Clock, Limits: deps.Limits}
	service := NewWorkflowActivationService(activation)
	if _, err := service.Activate(ctx, input); !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("queued run: %v", err)
	}
	now := deps.Clock.Now()
	claimed, err := run.Claim("lease", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvaluationRun(ctx, claimed, 1, fixture.Record(t, "claim", string(input.RunID), audit.ActionConversationEvaluationRunProgressed)); err != nil {
		t.Fatal(err)
	}
	run = claimed
	for i, revision := range run.Snapshot().Input.Cases {
		expected := revision.Snapshot().Definition.Settings().Expectations
		next, err := run.RecordCase("lease", i, model.EvaluationObservedOutcome{Kind: expected.Kind, ReferencedAssets: expected.ReferencedAssets, Locations: expected.Locations, Proposals: expected.Proposals}, 1, 0, now)
		if err != nil {
			t.Fatal(err)
		}
		record := fixture.Record(t, string(revision.Snapshot().ID), string(input.RunID), audit.ActionConversationEvaluationRunProgressed)
		if err := store.SaveEvaluationRun(ctx, next, run.Snapshot().Version, record); err != nil {
			t.Fatal(err)
		}
		run = next
	}
	for _, scenario := range []string{"wrong tenant", "missing run", "wrong suite", "configuration drift", "stale selection", "partial selection", "limits"} {
		t.Run(scenario, func(t *testing.T) {
			altered := input
			changed := activation
			switch scenario {
			case "wrong tenant":
				altered.TenantID = "other"
			case "missing run":
				altered.RunID = "missing"
			case "wrong suite":
				altered.Cases = altered.Cases[:1]
			case "configuration drift":
				values := append([]model.EvaluationRunProvider(nil), run.Snapshot().Input.Providers...)
				values[0].ConfigurationID = "changed"
				changed.Providers = evaluationProviderSnapshots{values: values}
			case "stale selection":
				altered.Expected = ports.WorkflowSelectionReference{WorkflowID: "other", RevisionID: "other"}
			case "partial selection":
				altered.Expected.WorkflowID = "other"
			case "limits":
				changed.Limits.Budget.ModelCalls = 1
			}
			if _, err := NewWorkflowActivationService(changed).Activate(ctx, altered); err == nil {
				t.Fatal("invalid activation accepted")
			}
			if _, found, _ := store.SelectedWorkflowRevision(ctx, input.TenantID); found {
				t.Fatal("invalid activation changed selection")
			}
		})
	}
	revision, err := service.Activate(ctx, input)
	if err != nil || revision.Snapshot().ID != input.RevisionID {
		t.Fatalf("activate: %v", err)
	}
	selected, found, err := store.SelectedWorkflowRevision(ctx, input.TenantID)
	if err != nil || !found || selected.RevisionID != input.RevisionID {
		t.Fatalf("selection not saved: %v", err)
	}
	records, err := store.ListTenantAuditRecords(ctx, input.TenantID, ports.AuditRecordPageRequest{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	activations := 0
	for _, record := range records {
		if record.Action == audit.ActionConversationWorkflowActivated {
			activations++
			if record.TargetID != string(input.WorkflowID) {
				t.Fatal("wrong audit target")
			}
		}
	}
	if activations != 1 {
		t.Fatalf("activation audits: %d", activations)
	}
	if _, err := service.Activate(ctx, input); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("stale selection: %v", err)
	}
}
