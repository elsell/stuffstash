package agentmodel

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestEvaluationWorkerRejectsHistoricalRuntimeBeforeProviderResolution(t *testing.T) {
	for _, scenario := range []string{"queued", "exhausted"} {
		t.Run(scenario, func(t *testing.T) {
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

			wantFailure := model.EvaluationRunFailureConfigurationChanged
			if scenario == "exhausted" {
				wantFailure = model.EvaluationRunFailureWorkerLost
				for attempt := 0; attempt < historical.Snapshot().Input.MaxAttempts; attempt++ {
					claimed, err := historical.Claim("old-worker-"+strconv.Itoa(attempt), clock.Now(), time.Minute)
					if err != nil {
						t.Fatal(err)
					}
					progress, ok := audit.NewRecord(audit.ID("historical-attempt-"+strconv.Itoa(attempt)), audit.TenantID(ref.TenantID), "", audit.PrincipalID(input.AuthorID), audit.ActionConversationEvaluationRunProgressed, audit.SourceSystem, audit.TargetConversationEvaluationRun, string(input.ID), clock.Now(), "", nil)
					if !ok {
						t.Fatal("invalid progress audit")
					}
					if err := store.SaveEvaluationRun(context.Background(), claimed, historical.Snapshot().Version, progress); err != nil {
						t.Fatal(err)
					}
					historical = claimed
					clock.now = clock.now.Add(time.Minute)
				}
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
			if result.Snapshot().FailureCode != wantFailure || executor.calls != 0 || providers.calls != 0 {
				t.Fatalf("historical run reached new runtime: state=%s code=%s modelCalls=%d resolutions=%d", result.Snapshot().State, result.Snapshot().FailureCode, executor.calls, providers.calls)
			}
		})
	}
}
