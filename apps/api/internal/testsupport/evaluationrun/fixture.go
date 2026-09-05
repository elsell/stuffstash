package evaluationrun

import (
	"strings"
	"testing"
	"time"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

const TenantID = "evaluation-run-home"

var Now = time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

func Run(t *testing.T, id string) model.EvaluationRun {
	t.Helper()
	limits := model.WorkflowLimits{Budget: model.WorkflowBudget{EvidenceRounds: 2, ModelCalls: 8, ElapsedSeconds: 60, FollowUpTurns: 2}, MaxStepAttempts: 2, MaxNameRunes: 100, MaxInstructionRunes: 4000}
	definition, err := model.NewWorkflowDefinition(model.WorkflowDefinitionInput{Name: "Household", Retrieval: model.WorkflowRetrievalExpanded, Response: model.WorkflowResponseGrounded, Budget: limits.Budget, Steps: []model.WorkflowStep{{Kind: model.WorkflowStepInterpret, Attempts: 1}, {Kind: model.WorkflowStepAssess, Attempts: 1}, {Kind: model.WorkflowStepRespond, Attempts: 1}}}, limits)
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := model.NewWorkflowRevision(model.WorkflowRevisionInput{ID: "workflow-revision", WorkflowID: "workflow", TenantID: TenantID, AuthorID: "owner", Number: 1, CreatedAt: Now, Definition: definition, Limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	var cases []model.EvaluationCaseRevision
	for _, id := range []string{"one", "two"} {
		definition, err := model.NewEvaluationCaseDefinition(model.EvaluationCaseDefinitionInput{Title: "Baby clothes", Utterance: "Where are my baby clothes?", Assets: []model.EvaluationFixtureAsset{{ID: "box", Title: "Attic box", Kind: model.EvaluationFixtureContainer}, {ID: "clothes", Title: "3–6 months", Kind: model.EvaluationFixtureItem, ParentID: "box", TagNames: []string{"baby", "clothes"}}}, Expectations: model.EvaluationExpectations{Kind: model.EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []model.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "box"}}}})
		if err != nil {
			t.Fatal(err)
		}
		revision, err := model.NewEvaluationCaseRevision(model.EvaluationCaseRevisionInput{ID: model.EvaluationCaseRevisionID(id + "-revision"), CaseID: model.EvaluationCaseID(id), TenantID: TenantID, AuthorID: "owner", Number: 1, CreatedAt: Now, Definition: definition})
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, revision)
	}
	run, err := model.NewEvaluationRun(model.EvaluationRunInput{ID: model.EvaluationRunID(id), TenantID: TenantID, AuthorID: "owner", CreatedAt: Now, Workflow: workflow, Cases: cases, Limits: limits, MaxAttempts: 2, Providers: []model.EvaluationRunProvider{{Step: model.WorkflowStepInterpret, ProfileID: "model", ConfigurationID: strings.Repeat("a", 64)}, {Step: model.WorkflowStepAssess, ProfileID: "model", ConfigurationID: strings.Repeat("a", 64)}}})
	if err != nil {
		t.Fatal(err)
	}
	return run
}
func Record(t *testing.T, id, runID string, action audit.Action) audit.Record {
	t.Helper()
	record, ok := audit.NewRecord(audit.ID(id), TenantID, "", "owner", action, audit.SourceSystem, "conversation_evaluation_run", runID, Now, "", nil)
	if !ok {
		t.Fatal("invalid fixture audit")
	}
	return record
}
