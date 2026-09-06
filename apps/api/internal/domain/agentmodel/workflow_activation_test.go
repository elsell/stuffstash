package agentmodel

import (
	"testing"
	"time"
)

func successfulActivationRun(t *testing.T) EvaluationRun {
	t.Helper()
	run, err := activationQueuedRun(t).Claim("lease", evaluationRunInput(t).CreatedAt, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for i, revision := range run.Snapshot().Input.Cases {
		expected := revision.Snapshot().Definition.Settings().Expectations
		observed := EvaluationObservedOutcome{Kind: expected.Kind, ReferencedAssets: expected.ReferencedAssets, Locations: expected.Locations, Proposals: expected.Proposals}
		run, err = run.RecordCase("lease", i, observed, 1, time.Second, evaluationRunInput(t).CreatedAt.Add(time.Duration(i+1)*time.Second))
		if err != nil {
			t.Fatal(err)
		}
	}
	return run
}
func TestWorkflowActivationRequiresExactSuccessfulEvidence(t *testing.T) {
	run := successfulActivationRun(t)
	pinned := run.Snapshot().Input
	candidate := WorkflowActivationCandidate{Workflow: pinned.Workflow, Limits: pinned.Limits, Providers: pinned.Providers}
	for _, revision := range pinned.Cases {
		value := revision.Snapshot()
		candidate.Cases = append(candidate.Cases, EvaluationCasePin{CaseID: value.CaseID, RevisionID: value.ID})
	}
	if err := run.ValidateActivation(candidate); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"queued", "failed", "missing case", "duplicate case", "wrong case revision", "wrong workflow", "changed configuration", "changed provider", "missing provider", "duplicate provider", "lower limits", "changed timestamp"} {
		t.Run(name, func(t *testing.T) {
			altered := candidate
			altered.Cases = append([]EvaluationCasePin(nil), candidate.Cases...)
			altered.Providers = append([]EvaluationRunProvider(nil), candidate.Providers...)
			evidence := run
			switch name {
			case "queued":
				evidence = activationQueuedRun(t)
			case "failed":
				failed, _ := activationQueuedRun(t).Claim("lease", evaluationRunInput(t).CreatedAt, time.Minute)
				for i := range 2 {
					failed, _ = failed.RecordCase("lease", i, EvaluationObservedOutcome{Kind: EvaluationOutcomeFailure}, 1, time.Second, evaluationRunInput(t).CreatedAt.Add(time.Duration(i+1)*time.Second))
				}
				evidence = failed
			case "missing case":
				altered.Cases = altered.Cases[:1]
			case "duplicate case":
				altered.Cases[1] = altered.Cases[0]
			case "wrong case revision":
				altered.Cases[0].RevisionID = "another"
			case "wrong workflow":
				value := pinned.Workflow.Snapshot()
				value.ID = "another"
				altered.Workflow, _ = NewWorkflowRevision(value)
			case "changed configuration":
				altered.Providers[0].ConfigurationID = "changed"
			case "changed provider":
				altered.Providers[0].ProfileID = "other"
			case "missing provider":
				altered.Providers = nil
			case "duplicate provider":
				altered.Providers = append(altered.Providers, altered.Providers[0])
			case "changed timestamp":
				value := pinned.Workflow.Snapshot()
				value.CreatedAt = value.CreatedAt.Add(time.Nanosecond)
				altered.Workflow, _ = NewWorkflowRevision(value)
			case "lower limits":
				altered.Limits.Budget.ModelCalls = 1
			}
			if err := evidence.ValidateActivation(altered); err == nil {
				t.Fatal("invalid evidence accepted")
			}
		})
	}
	equivalent := pinned.Workflow.Snapshot()
	equivalent.CreatedAt = equivalent.CreatedAt.In(time.FixedZone("offset", 3600))
	candidate.Workflow, _ = NewWorkflowRevision(equivalent)
	if err := run.ValidateActivation(candidate); err != nil {
		t.Fatalf("equivalent timestamp rejected: %v", err)
	}
	candidate.Cases[0], candidate.Cases[1] = candidate.Cases[1], candidate.Cases[0]
	if err := run.ValidateActivation(candidate); err != nil {
		t.Fatalf("ordering should not change evidence: %v", err)
	}
}

func activationQueuedRun(t *testing.T) EvaluationRun {
	t.Helper()
	run, err := NewEvaluationRun(evaluationRunInput(t))
	if err != nil {
		t.Fatal(err)
	}
	return run
}
