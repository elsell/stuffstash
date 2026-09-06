package agentmodel

import (
	"testing"
	"time"
)

func workflowRevisionTestInput(t *testing.T) WorkflowRevisionInput {
	t.Helper()
	definition, err := NewWorkflowDefinition(workflowTestInput(), workflowTestLimits())
	if err != nil {
		t.Fatal(err)
	}
	return WorkflowRevisionInput{
		ID: "revision-1", WorkflowID: "workflow-1", TenantID: "tenant-1", AuthorID: "owner-1",
		Number: 1, Definition: definition, Limits: workflowTestLimits(), CreatedAt: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
	}
}

func TestWorkflowRevisionPreservesIdentityAndDefinitionSnapshot(t *testing.T) {
	input := workflowRevisionTestInput(t)
	revision, err := NewWorkflowRevision(input)
	if err != nil {
		t.Fatal(err)
	}
	input.ID = "changed"
	result := revision.Snapshot()
	if result.ID != "revision-1" || result.WorkflowID != "workflow-1" || result.TenantID != "tenant-1" || result.Number != 1 || result.CreatedAt != input.CreatedAt {
		t.Fatalf("revision identity changed: %+v", result)
	}
	settings := result.Definition.Settings()
	settings.Instructions = "mutated"
	if revision.Snapshot().Definition.Settings().Instructions != "" {
		t.Fatal("revision definition mutated")
	}
}

func TestWorkflowRevisionRejectsInvalidSnapshots(t *testing.T) {
	cases := map[string]func(*WorkflowRevisionInput){
		"missing revision":             func(v *WorkflowRevisionInput) { v.ID = "" },
		"missing workflow":             func(v *WorkflowRevisionInput) { v.WorkflowID = "" },
		"missing tenant":               func(v *WorkflowRevisionInput) { v.TenantID = "" },
		"missing author":               func(v *WorkflowRevisionInput) { v.AuthorID = "" },
		"path identifier":              func(v *WorkflowRevisionInput) { v.ID = "../another" },
		"blank tenant":                 func(v *WorkflowRevisionInput) { v.TenantID = " " },
		"missing sequence":             func(v *WorkflowRevisionInput) { v.Number = 0 },
		"missing timestamp":            func(v *WorkflowRevisionInput) { v.CreatedAt = time.Time{} },
		"missing definition":           func(v *WorkflowRevisionInput) { v.Definition = WorkflowDefinition{} },
		"incompatible recorded limits": func(v *WorkflowRevisionInput) { v.Limits.Budget.ModelCalls = 1 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			input := workflowRevisionTestInput(t)
			change(&input)
			if _, err := NewWorkflowRevision(input); err == nil {
				t.Fatal("invalid revision accepted")
			}
		})
	}
}

func TestWorkflowRevisionPreservesExistingTenantIdentity(t *testing.T) {
	input := workflowRevisionTestInput(t)
	input.TenantID = "tenant.home"
	revision, err := NewWorkflowRevision(input)
	if err != nil || revision.Snapshot().TenantID != input.TenantID {
		t.Fatalf("valid tenant reference changed or rejected: %v", err)
	}
}
