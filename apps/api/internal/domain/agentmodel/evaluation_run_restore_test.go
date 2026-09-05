package agentmodel

import (
	"testing"
	"time"
)

func TestEvaluationRunRestoresValidatedProgress(t *testing.T) {
	input := evaluationRunInput(t)
	now := input.CreatedAt
	queued, _ := NewEvaluationRun(input)
	running, _ := queued.Claim("worker", now, time.Minute)
	partial, err := running.RecordCase("worker", 0, EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}}, 1, time.Second, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	completed, err := partial.RecordCase("worker", 1, EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}}, 1, time.Second, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	cancelled, _ := partial.Cancel(now.Add(2 * time.Second))
	failed, _ := partial.Fail("worker", EvaluationRunFailureExecution, now.Add(2*time.Second))
	for _, original := range []EvaluationRun{queued, running, partial, completed, cancelled, failed} {
		restored, err := RestoreEvaluationRun(original.Snapshot())
		if err != nil || restored.Snapshot().State != original.Snapshot().State || restored.Snapshot().Version != original.Snapshot().Version {
			t.Fatalf("restore %s: %v", original.Snapshot().State, err)
		}
	}
	saved := partial.Snapshot()
	restored, err := RestoreEvaluationRun(saved)
	if err != nil {
		t.Fatal(err)
	}
	saved.Results[0].Observation.ReferencedAssets[0] = "changed"
	saved.Input.Providers[0].ProfileID = "changed"
	if restored.Snapshot().Results[0].Observation.ReferencedAssets[0] != "clothes" || restored.Snapshot().Input.Providers[0].ProfileID != "model" {
		t.Fatal("rehydration retained caller collections")
	}
	for name, change := range map[string]func(*EvaluationRunSnapshot){
		"forged verdict":        func(v *EvaluationRunSnapshot) { v.Results[0].Verdict.Passed = false },
		"foreign case":          func(v *EvaluationRunSnapshot) { v.Results[0].CaseRevisionID = "outside" },
		"unknown observed item": func(v *EvaluationRunSnapshot) { v.Results[0].Observation.ReferencedAssets[0] = "outside" },
		"missing lease":         func(v *EvaluationRunSnapshot) { v.LeaseToken = "" },
		"expired at update":     func(v *EvaluationRunSnapshot) { v.LeaseUntil = v.UpdatedAt },
		"early result":          func(v *EvaluationRunSnapshot) { v.Results[0].CompletedAt = now.Add(-time.Second) },
		"negative calls":        func(v *EvaluationRunSnapshot) { v.Results[0].ModelCalls = -1 },
		"unknown state":         func(v *EvaluationRunSnapshot) { v.State = "unknown" },
		"queued with results":   func(v *EvaluationRunSnapshot) { v.State = EvaluationRunQueued },
		"premature success": func(v *EvaluationRunSnapshot) {
			v.State = EvaluationRunSucceeded
			v.LeaseToken = ""
			v.LeaseUntil = time.Time{}
			v.FinishedAt = v.UpdatedAt
		},
		"version below transitions": func(v *EvaluationRunSnapshot) { v.Version = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			snapshot := partial.Snapshot()
			change(&snapshot)
			if _, err := RestoreEvaluationRun(snapshot); err == nil {
				t.Fatal("corrupt persisted run accepted")
			}
		})
	}
}
