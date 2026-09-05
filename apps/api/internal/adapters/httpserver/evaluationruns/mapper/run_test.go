package mapper

import (
	"encoding/json"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
	"strings"
	"testing"
	"time"
)

func TestRunResponsePreservesResultsWithoutWorkerSecrets(t *testing.T) {
	run := evaluationrun.Run(t, "run")
	run, err := run.Claim("private-worker-token", evaluationrun.Now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	running := RunToResponse(run)
	wire, err := json.Marshal(running)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private-worker-token", "leaseToken", "leaseUntil", "maxAttempts", "inputJSON", "progressJSON"} {
		if strings.Contains(string(wire), forbidden) {
			t.Fatalf("private field exposed: %s", forbidden)
		}
	}
	if running.StartedAt == nil || running.FinishedAt != nil || running.Coverage != "text_only" {
		t.Fatalf("running lifecycle: %+v", running)
	}
	observation := model.EvaluationObservedOutcome{Kind: model.EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []model.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "box"}}}
	run, err = run.RecordCase("private-worker-token", 0, observation, 2, 1500*time.Millisecond, evaluationrun.Now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	// Second case expects a proposal; a grounded answer must remain a failed case.
	run, err = run.RecordCase("private-worker-token", 1, observation, 3, time.Second, evaluationrun.Now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	response := RunToResponse(run)
	if response.State != "failed" || response.CompletedCases != 2 || response.PassedCases != 1 || response.FinishedAt == nil || len(response.Results) != 2 {
		t.Fatalf("finished lifecycle: %+v", response)
	}
	first, second := response.Results[0], response.Results[1]
	if first.CaseRevisionID != "one-revision" || !first.Verdict.Passed || first.DurationMilliseconds != 1500 || first.ModelCalls != 2 || first.Observation.Locations[0].AncestorID != "box" {
		t.Fatalf("result lost: %+v", first)
	}
	if second.Verdict.Passed || len(second.Verdict.Failures) != 2 || second.Verdict.Failures[0].Code != "unexpected_outcome" || second.Verdict.Failures[1].Code != "missing_proposal" || second.Verdict.Failures[1].FixtureID != "clothes" || second.Verdict.Failures[1].Operation != "checkout" {
		t.Fatalf("verdict lost: %+v", second)
	}
}
