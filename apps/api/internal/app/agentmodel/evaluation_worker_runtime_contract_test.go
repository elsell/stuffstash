package agentmodel

import (
	"context"
	"encoding/json"
	"testing"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestEvaluationWorkerRejectsHistoricalRuntimeBeforeProviderResolution(t *testing.T) {
	deps, store, executor, clock := evaluationWorkerSetup(t)
	ref := workerReference(t, store)
	current, found, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
	if err != nil || !found {
		t.Fatalf("load queued run: %v", err)
	}
	snapshot := current.Snapshot()
	raw, err := json.Marshal(snapshot.Input)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	delete(fields, "RuntimeContract")
	fields["ID"] = json.RawMessage(`"historical-worker-run"`)
	raw, err = json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	var input model.EvaluationRunInput
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatal(err)
	}
	input.Workflow, input.Cases = snapshot.Input.Workflow, snapshot.Input.Cases
	snapshot.Input = input
	historical, err := model.RestoreEvaluationRun(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	record, ok := audit.NewRecord("historical-worker-audit", audit.TenantID(ref.TenantID), "", audit.PrincipalID(input.AuthorID), audit.ActionConversationEvaluationRunCreated, audit.SourceSystem, audit.TargetConversationEvaluationRun, string(input.ID), clock.Now(), "", nil)
	if !ok {
		t.Fatal("invalid audit fixture")
	}
	if err := store.SaveEvaluationRun(context.Background(), historical, 0, record); err != nil {
		t.Fatal(err)
	}
	historicalRef := ports.EvaluationRunReference{TenantID: ref.TenantID, ID: input.ID}
	if err := NewEvaluationWorker(deps).Process(context.Background(), historicalRef); err != nil {
		t.Fatal(err)
	}
	result, found, err := store.EvaluationRun(context.Background(), ref.TenantID, input.ID)
	if err != nil || !found {
		t.Fatalf("load result: %v", err)
	}
	providers := deps.Providers.(*evaluationWorkerProviders)
	if result.Snapshot().FailureCode != model.EvaluationRunFailureConfigurationChanged || executor.calls != 0 || providers.calls != 0 {
		t.Fatalf("historical run reached new runtime: state=%s code=%s modelCalls=%d resolutions=%d", result.Snapshot().State, result.Snapshot().FailureCode, executor.calls, providers.calls)
	}
}
