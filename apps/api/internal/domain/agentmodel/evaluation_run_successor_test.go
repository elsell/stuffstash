package agentmodel

import (
	"testing"
	"time"
)

func TestEvaluationRunSuccessorsPreserveHistoryAndTerminalStates(t *testing.T) {
	input := evaluationRunInput(t)
	now := input.CreatedAt
	queued, _ := NewEvaluationRun(input)
	running, _ := queued.Claim("worker", now, time.Minute)
	partial, _ := running.RecordCase("worker", 0, EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}}, 1, time.Second, now.Add(time.Second))
	renewed, _ := partial.Renew("worker", now.Add(30*time.Second), time.Minute)
	recovered, _ := partial.Claim("next", now.Add(time.Minute), time.Minute)
	cancelled, _ := partial.Cancel(now.Add(2 * time.Second))
	failed, _ := partial.Fail("worker", EvaluationRunFailureExecution, now.Add(2*time.Second))
	completed, _ := partial.RecordCase("worker", 1, EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}}, 1, time.Second, now.Add(2*time.Second))
	for _, pair := range [][2]EvaluationRun{{queued, running}, {running, partial}, {partial, renewed}, {partial, recovered}, {partial, cancelled}, {partial, failed}, {partial, completed}} {
		if !pair[1].IsSuccessorOf(pair[0]) {
			t.Fatalf("valid successor rejected: %s -> %s", pair[0].Snapshot().State, pair[1].Snapshot().State)
		}
	}
	for name, change := range map[string]func(*EvaluationRunSnapshot){
		"rewrite case": func(v *EvaluationRunSnapshot) {
			v.Input.Cases[0], v.Input.Cases[1] = v.Input.Cases[1], v.Input.Cases[0]
		},
		"change provider": func(v *EvaluationRunSnapshot) {
			for i := range v.Input.Providers {
				v.Input.Providers[i].ProfileID = "different"
			}
		},
		"change author": func(v *EvaluationRunSnapshot) { v.Input.AuthorID = "another" },
		"skip version":  func(v *EvaluationRunSnapshot) { v.Version++ },
	} {
		t.Run(name, func(t *testing.T) {
			value := running.Snapshot()
			change(&value)
			forged, err := RestoreEvaluationRun(value)
			if err != nil {
				t.Fatal(err)
			}
			if forged.IsSuccessorOf(queued) {
				t.Fatal("fabricated successor accepted")
			}
		})
	}
	// This is independently valid persisted state, but it rewrites a completed trial.
	altered := completed.Snapshot()
	altered.Results[0].ModelCalls++
	forged, err := RestoreEvaluationRun(altered)
	if err != nil {
		t.Fatal(err)
	}
	if forged.IsSuccessorOf(partial) {
		t.Fatal("completed case metrics overwritten")
	}
	resurrected := partial.Snapshot()
	resurrected.Version = cancelled.Snapshot().Version + 1
	resurrected.UpdatedAt = cancelled.Snapshot().UpdatedAt
	forged, err = RestoreEvaluationRun(resurrected)
	if err != nil {
		t.Fatal(err)
	}
	if forged.IsSuccessorOf(cancelled) {
		t.Fatal("cancelled run revived")
	}
	if completed.IsSuccessorOf(running) {
		t.Fatal("multiple transitions collapsed into one CAS")
	}
}
