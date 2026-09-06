package agentmodel

import (
	"strings"
	"testing"
	"time"
)

func evaluationRunInput(t *testing.T) EvaluationRunInput {
	t.Helper()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	limits := workflowTestLimits()
	definition, err := NewWorkflowDefinition(workflowTestInput(), limits)
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := NewWorkflowRevision(WorkflowRevisionInput{ID: "revision", WorkflowID: "workflow", TenantID: "home", AuthorID: "owner", Number: 1, Definition: definition, Limits: limits, CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	fixture := fixtureEvaluationInput()
	fixture.Expectations.Locations = nil
	def, err := NewEvaluationCaseDefinition(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var cases []EvaluationCaseRevision
	for _, id := range []string{"one", "two"} {
		revision, err := NewEvaluationCaseRevision(EvaluationCaseRevisionInput{ID: EvaluationCaseRevisionID("revision-" + id), CaseID: EvaluationCaseID(id), TenantID: "home", AuthorID: "owner", Number: 1, Definition: def, CreatedAt: now})
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, revision)
	}
	return EvaluationRunInput{ID: "run", TenantID: "home", AuthorID: "owner", CreatedAt: now, Workflow: workflow, Cases: cases, Limits: limits, MaxAttempts: 2, Providers: []EvaluationRunProvider{{ProfileID: "model", ConfigurationID: strings.Repeat("a", 64)}}}
}

func TestEvaluationRunPinsSnapshotsAndRejectsInvalidInputs(t *testing.T) {
	input := evaluationRunInput(t)
	run, err := NewEvaluationRun(input)
	if err != nil {
		t.Fatal(err)
	}
	input.Cases[0] = EvaluationCaseRevision{}
	input.Providers[0].ProfileID = "changed"
	snapshot := run.Snapshot()
	if snapshot.Input.RuntimeContract != CurrentEvaluationRuntimeContract || snapshot.State != EvaluationRunQueued || snapshot.Version != 1 || snapshot.Input.Cases[0].Snapshot().CaseID != "one" || snapshot.Input.Providers[0].ProfileID != "model" {
		t.Fatal("run did not own queued snapshot")
	}
	snapshot.Input.Cases[0] = EvaluationCaseRevision{}
	snapshot.Input.Providers[0].ProfileID = "changed"
	if run.Snapshot().Input.Cases[0].Snapshot().CaseID != "one" || run.Snapshot().Input.Providers[0].ProfileID != "model" {
		t.Fatal("snapshot mutation reached run")
	}
	for name, change := range map[string]func(*EvaluationRunInput){
		"legacy runtime":          func(v *EvaluationRunInput) { v.RuntimeContract = LegacyEvaluationRuntimeContract },
		"unknown runtime":         func(v *EvaluationRunInput) { v.RuntimeContract = "unknown-runtime" },
		"no cases":                func(v *EvaluationRunInput) { v.Cases = nil },
		"duplicate case":          func(v *EvaluationRunInput) { v.Cases[1] = v.Cases[0] },
		"wrong tenant":            func(v *EvaluationRunInput) { v.TenantID = "other" },
		"missing binding":         func(v *EvaluationRunInput) { v.Providers = nil },
		"duplicate binding":       func(v *EvaluationRunInput) { v.Providers = append(v.Providers, v.Providers[0]) },
		"arbitrary configuration": func(v *EvaluationRunInput) { v.Providers[0].ConfigurationID = "secret" },
		"no attempts":             func(v *EvaluationRunInput) { v.MaxAttempts = 0 },
		"inconsistent profile": func(v *EvaluationRunInput) {
			v.Providers = append(v.Providers, EvaluationRunProvider{ProfileID: "model", ConfigurationID: strings.Repeat("b", 64)})
		},
		"inconsistent default": func(v *EvaluationRunInput) {
			v.Providers = append(v.Providers, EvaluationRunProvider{ProfileID: "other-model", ConfigurationID: strings.Repeat("a", 64)})
		},
		"invalid limits": func(v *EvaluationRunInput) { v.Limits.Budget.ToolCalls = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			input := evaluationRunInput(t)
			change(&input)
			if _, err := NewEvaluationRun(input); err == nil {
				t.Fatal("invalid run accepted")
			}
		})
	}
}

func TestEvaluationRunLeaseRecoveryFencesStaleOwnersAndPreservesProgress(t *testing.T) {
	input := evaluationRunInput(t)
	now := input.CreatedAt
	run, err := NewEvaluationRun(input)
	if err != nil {
		t.Fatal(err)
	}
	run, err = run.Claim("first", now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := run.Claim("second", now, time.Minute); err == nil {
		t.Fatal("active lease stolen")
	}
	observation := EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}}
	run, err = run.RecordCase("first", 0, observation, 2, time.Second, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := run.RecordCase("first", 0, observation, 2, time.Second, now.Add(2*time.Second)); err == nil {
		t.Fatal("completed case replaced")
	}
	recovered, err := run.Claim("second", now.Add(time.Minute), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Snapshot().Attempts != 2 || len(recovered.Snapshot().Results) != 1 {
		t.Fatal("recovery lost completed work")
	}
	if _, err := recovered.RecordCase("first", 1, observation, 1, time.Second, now.Add(time.Minute)); err == nil {
		t.Fatal("stale owner recorded result")
	}
	if _, err := recovered.Renew("first", now.Add(time.Minute), time.Minute); err == nil {
		t.Fatal("stale owner renewed")
	}
	recovered, err = recovered.RecordCase("second", 1, observation, 1, time.Second, now.Add(61*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	final := recovered.Snapshot()
	if final.State != EvaluationRunSucceeded || final.LeaseToken != "" || !final.LeaseUntil.IsZero() || final.FinishedAt.IsZero() || final.Version != 5 {
		t.Fatalf("bad terminal state: %+v", final)
	}
	final.Results[0].Observation.ReferencedAssets[0] = "unknown"
	if recovered.Snapshot().Results[0].Observation.ReferencedAssets[0] != "clothes" {
		t.Fatal("result mutated run")
	}
	if _, err := recovered.Claim("third", now.Add(time.Hour), time.Minute); err == nil {
		t.Fatal("terminal run reclaimed")
	}
}

func TestEvaluationRunJudgesOutcomesAndCancellation(t *testing.T) {
	input := evaluationRunInput(t)
	now := input.CreatedAt
	run, _ := NewEvaluationRun(input)
	run, _ = run.Claim("worker", now, time.Minute)
	if _, err := run.RecordCase("worker", 0, EvaluationObservedOutcome{Kind: "garbage"}, 0, 0, now); err == nil {
		t.Fatal("invalid observation accepted")
	}
	if _, err := run.RecordCase("worker", 1, EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer}, 0, 0, now); err == nil {
		t.Fatal("out of order result accepted")
	}
	var err error
	run, err = run.RecordCase("worker", 0, EvaluationObservedOutcome{Kind: EvaluationOutcomeClarification}, 1, time.Second, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if run.Snapshot().State != EvaluationRunRunning || run.Snapshot().Results[0].Verdict.Passed {
		t.Fatal("failed assertion ended suite or passed")
	}
	cancelled, err := run.Cancel(now.Add(2 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Snapshot().State != EvaluationRunCancelled || cancelled.Snapshot().LeaseToken != "" {
		t.Fatal("cancellation retained lease")
	}
	repeated, err := cancelled.Cancel(now.Add(3 * time.Second))
	if err != nil || repeated.Snapshot().Version != cancelled.Snapshot().Version {
		t.Fatal("cancel not idempotent")
	}
	if _, err := cancelled.RecordCase("worker", 1, EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer}, 1, time.Second, now.Add(3*time.Second)); err == nil {
		t.Fatal("cancelled worker wrote outcome")
	}
	finished, err := run.RecordCase("worker", 1, EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}}, 1, time.Second, now.Add(2*time.Second))
	if err != nil || finished.Snapshot().State != EvaluationRunFailed {
		t.Fatalf("failed assertion lost: %v", err)
	}
}

func TestEvaluationRunRejectsExpiredLateAndExhaustedWork(t *testing.T) {
	input := evaluationRunInput(t)
	now := input.CreatedAt
	input.MaxAttempts = 1
	run, _ := NewEvaluationRun(input)
	run, _ = run.Claim("worker", now, time.Minute)
	if _, err := run.Renew("worker", now.Add(time.Minute), time.Minute); err == nil {
		t.Fatal("expired lease revived")
	}
	if _, err := run.RecordCase("worker", 0, EvaluationObservedOutcome{Kind: EvaluationOutcomeAnswer}, 1, time.Second, now.Add(time.Minute)); err == nil {
		t.Fatal("expired result accepted")
	}
	if _, err := run.Claim("next", now.Add(time.Minute), time.Minute); err == nil {
		t.Fatal("claim budget ignored")
	}
	if _, err := run.FailExpired(now); err == nil {
		t.Fatal("live lease failed by recovery")
	}
	failed, err := run.FailExpired(now.Add(time.Minute))
	if err != nil || failed.Snapshot().State != EvaluationRunFailed || failed.Snapshot().FailureCode != EvaluationRunFailureWorkerLost {
		t.Fatalf("lost worker not terminated: %v", err)
	}
	if _, err := run.Cancel(now.Add(-time.Second)); err == nil {
		t.Fatal("clock regression accepted")
	}
}

func TestEvaluationRunRenewalAndSafeFailures(t *testing.T) {
	input := evaluationRunInput(t)
	now := input.CreatedAt
	run, _ := NewEvaluationRun(input)
	run, _ = run.Claim("worker", now, time.Minute)
	renewed, err := run.Renew("worker", now.Add(30*time.Second), time.Minute)
	if err != nil || !renewed.Snapshot().LeaseUntil.Equal(now.Add(90*time.Second)) || renewed.Snapshot().Version != 3 {
		t.Fatalf("renewal: %v", err)
	}
	if _, err := renewed.Fail("worker", EvaluationRunFailureCode("secret payload"), now.Add(31*time.Second)); err == nil {
		t.Fatal("unsafe failure code accepted")
	}
	if _, err := renewed.Fail("other", EvaluationRunFailureExecution, now.Add(31*time.Second)); err == nil {
		t.Fatal("wrong owner failed run")
	}
	failed, err := renewed.Fail("worker", EvaluationRunFailureConfigurationChanged, now.Add(31*time.Second))
	if err != nil || failed.Snapshot().FailureCode != EvaluationRunFailureConfigurationChanged || failed.Snapshot().State != EvaluationRunFailed || failed.Snapshot().LeaseToken != "" {
		t.Fatalf("safe failure: %v", err)
	}
	queued, _ := NewEvaluationRun(input)
	cancelled, err := queued.Cancel(now)
	if err != nil || cancelled.Snapshot().State != EvaluationRunCancelled || !cancelled.Snapshot().StartedAt.IsZero() {
		t.Fatal("queued cancellation invalid")
	}
}

func TestEvaluationRunResolvesOnlyRequiredPinnedProviders(t *testing.T) {
	input := evaluationRunInput(t)
	revision := input.Workflow.Snapshot()
	settings := revision.Definition.Settings()
	settings.ProviderProfileID = "explicit"
	definition, err := NewWorkflowDefinition(settings, input.Limits)
	if err != nil {
		t.Fatal(err)
	}
	revision.Definition = definition
	input.Workflow, err = NewWorkflowRevision(revision)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewEvaluationRun(input); err == nil {
		t.Fatal("explicit profile silently substituted")
	}
	input.Providers[0].ProfileID = "explicit"
	if _, err := NewEvaluationRun(input); err != nil {
		t.Fatalf("conversation rejected chosen model: %v", err)
	}
	caseSnapshot := input.Cases[0].Snapshot()
	caseSnapshot.TenantID = "other"
	input.Cases[0], err = NewEvaluationCaseRevision(caseSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewEvaluationRun(input); err == nil {
		t.Fatal("cross-tenant case accepted")
	}
}
