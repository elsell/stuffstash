package gormstore

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHistoricalWorkflowSettingsConvertToModelLoop(t *testing.T) {
	const historicalJSON = `{"Definition":{"Name":"Home voice","Retrieval":"expanded","Response":"generated_with_grounded_fallback","Budget":{"EvidenceRounds":2,"ModelCalls":8,"ElapsedSeconds":60,"FollowUpTurns":4},"Steps":[{"Kind":"interpret","ProviderProfileID":"chosen-model","Instructions":"Prefer household tags.","Attempts":2},{"Kind":"assess","ProviderProfileID":"obsolete-assessor","Instructions":"Require resolved status.","Attempts":2},{"Kind":"respond","ProviderProfileID":"obsolete-wording","Instructions":"Use prescribed wording.","Attempts":1}]},"Limits":{"Budget":{"EvidenceRounds":8,"ModelCalls":12,"ElapsedSeconds":120,"FollowUpTurns":8},"MaxStepAttempts":3,"MaxNameRunes":100,"MaxInstructionRunes":4000}}`
	created := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	stored := conversationWorkflowRevisionModel{TenantID: "home", WorkflowID: "workflow", ID: "old-revision", Number: 3, AuthorID: "owner", CreatedAt: created, SnapshotJSON: historicalJSON}
	revision, err := workflowRevisionFromModel(stored)
	if err != nil {
		t.Fatalf("historical settings unreadable: %v", err)
	}
	snapshot := revision.Snapshot()
	if string(snapshot.ID) != stored.ID || snapshot.Number != 3 || string(snapshot.AuthorID) != "owner" || !snapshot.CreatedAt.Equal(created) {
		t.Fatal("migration changed revision identity or attribution")
	}
	raw, err := json.Marshal(snapshot.Definition.Settings())
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(raw, &settings); err != nil {
		t.Fatal(err)
	}
	if settings["ProviderProfileID"] != "chosen-model" || settings["Instructions"] != "Prefer household tags." {
		t.Fatalf("chosen model or useful guidance lost: %s", raw)
	}
	for _, retired := range []string{"Steps", "Retrieval", "Response"} {
		if _, exists := settings[retired]; exists {
			t.Fatalf("retired setting survives migration: %s", retired)
		}
	}
	budget, ok := settings["Budget"].(map[string]any)
	if !ok || budget["ToolCalls"] != float64(6) || budget["ModelCalls"] != float64(8) || budget["ElapsedSeconds"] != float64(60) || budget["FollowUpTurns"] != float64(4) {
		t.Fatalf("useful budgets not preserved: %+v", budget)
	}
	raw, err = json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata["SettingsMigration"] != "legacy-investigation-v1" {
		t.Fatal("historical conversion was not disclosed")
	}
	// Re-encoding the converted snapshot preserves its provenance and never
	// reintroduces the old executable shape.
	converted, err := workflowRevisionModel(revision)
	if err != nil {
		t.Fatal(err)
	}
	again, err := workflowRevisionFromModel(converted)
	if err != nil {
		t.Fatal(err)
	}

	roundTrip := again.Snapshot()
	if roundTrip.TenantID != snapshot.TenantID || roundTrip.WorkflowID != snapshot.WorkflowID || roundTrip.ID != snapshot.ID || roundTrip.AuthorID != snapshot.AuthorID || roundTrip.Number != snapshot.Number || !roundTrip.CreatedAt.Equal(snapshot.CreatedAt) {
		t.Fatal("round-trip changed historical identity")
	}
	raw, err = json.Marshal(roundTrip)
	if err != nil {
		t.Fatal(err)
	}
	metadata = map[string]any{}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata["SettingsMigration"] != "legacy-investigation-v1" {
		t.Fatal("round-trip erased migration provenance")
	}
	next, err := json.Marshal(again.Snapshot().Definition.Settings())
	if err != nil {
		t.Fatal(err)
	}
	original, err := json.Marshal(snapshot.Definition.Settings())
	if err != nil || string(next) != string(original) {
		t.Fatal("converted settings did not survive persistence")
	}
}
