package gormstore

import (
	"encoding/json"
	"reflect"
	"testing"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

func TestHistoricalEvaluationWorkflowConversionKeepsLegacyEvidence(t *testing.T) {
	run := fixture.Run(t, "historical")
	inputJSON, progressJSON, err := encodeEvaluationRun(run)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal([]byte(inputJSON), &document); err != nil {
		t.Fatal(err)
	}
	delete(document, "RuntimeContract")
	workflow := document["Workflow"].(map[string]any)
	definition := workflow["Definition"].(map[string]any)
	definition["Retrieval"] = "expanded"
	definition["Response"] = "grounded"
	definition["Steps"] = []map[string]any{
		{"Kind": "interpret", "ProviderProfileID": "model", "Instructions": "Use tags.", "Attempts": 1},
		{"Kind": "assess", "Attempts": 1},
		{"Kind": "respond", "Attempts": 1},
	}
	budget := definition["Budget"].(map[string]any)
	delete(budget, "ToolCalls")
	budget["EvidenceRounds"] = 2
	for _, limits := range []map[string]any{document["Limits"].(map[string]any), workflow["Limits"].(map[string]any)} {
		limits["MaxStepAttempts"] = 2
		b := limits["Budget"].(map[string]any)
		delete(b, "ToolCalls")
		b["EvidenceRounds"] = 2
	}
	providers := document["Providers"].([]any)
	pin := providers[0].(map[string]any)
	pin["Step"] = "interpret"
	second := map[string]any{"Step": "assess", "ProfileID": pin["ProfileID"], "ConfigurationID": pin["ConfigurationID"]}
	document["Providers"] = append(providers, second)
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := decodeEvaluationRun(string(raw), progressJSON)
	if err != nil {
		t.Fatalf("old evaluation unreadable: %v", err)
	}
	snapshot := restored.Snapshot()
	revision := snapshot.Input.Workflow.Snapshot()
	settings := revision.Definition.Settings()
	if snapshot.Input.RuntimeContract != model.LegacyEvaluationRuntimeContract || revision.SettingsMigration != model.WorkflowSettingsMigrationLegacy || settings.ProviderProfileID != "model" || settings.Instructions != "Use tags." || settings.Budget.ToolCalls != 6 || len(snapshot.Input.Providers) != 2 {
		t.Fatalf("historical settings or evidence provenance lost: %+v", snapshot.Input)
	}
	if _, err := model.NewEvaluationRun(snapshot.Input); err == nil {
		t.Fatal("historical data became executable")
	}
	encoded, progress, err := encodeEvaluationRun(restored)
	if err != nil {
		t.Fatal(err)
	}
	again, err := decodeEvaluationRun(encoded, progress)
	if err != nil || again.Snapshot().Input.RuntimeContract != model.LegacyEvaluationRuntimeContract || again.Snapshot().Input.Workflow.Snapshot().SettingsMigration != model.WorkflowSettingsMigrationLegacy {
		t.Fatalf("roundtrip erased provenance: %v", err)
	}
}

func TestCurrentEvaluationOnConvertedWorkflowPreservesOperatorLimits(t *testing.T) {
	input := fixture.Run(t, "current-converted").Snapshot().Input
	revision := input.Workflow.Snapshot()
	revision.SettingsMigration = model.WorkflowSettingsMigrationLegacy
	var err error
	input.Workflow, err = model.NewWorkflowRevision(revision)
	if err != nil {
		t.Fatal(err)
	}
	input.Limits.Budget.ToolCalls = 12
	run, err := model.NewEvaluationRun(input)
	if err != nil {
		t.Fatal(err)
	}
	encoded, progress, err := encodeEvaluationRun(run)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := decodeEvaluationRun(encoded, progress)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(run.Snapshot().Input, restored.Snapshot().Input) {
		t.Fatal("persistence changed immutable current-run input because its workflow was migrated")
	}
}
