package agentmodel

import (
	"strings"
	"testing"
)

func workflowTestInput() WorkflowDefinitionInput {
	return WorkflowDefinitionInput{
		Name:   "Household voice",
		Budget: WorkflowBudget{ToolCalls: 3, ModelCalls: 8, ElapsedSeconds: 60, FollowUpTurns: 6},
	}
}

func workflowTestLimits() WorkflowLimits {
	return WorkflowLimits{Budget: WorkflowBudget{ToolCalls: 6, ModelCalls: 20, ElapsedSeconds: 180, FollowUpTurns: 12}, MaxNameRunes: 120, MaxInstructionRunes: 4000}
}

func TestWorkflowDefinitionAcceptsTuningWithinOperatorLimits(t *testing.T) {
	input := workflowTestInput()
	input.Instructions = "Prefer household tags when names differ."
	definition, err := NewWorkflowDefinition(input, workflowTestLimits())
	if err != nil {
		t.Fatal(err)
	}
	if definition.Settings().Instructions != input.Instructions || definition.Settings().Budget != input.Budget {
		t.Fatal("workflow tuning was discarded")
	}
}

func TestWorkflowDefinitionRejectsInvalidPolicy(t *testing.T) {
	cases := map[string]func(*WorkflowDefinitionInput){
		"empty name":          func(v *WorkflowDefinitionInput) { v.Name = " " },
		"zero tool budget":    func(v *WorkflowDefinitionInput) { v.Budget.ToolCalls = 0 },
		"excessive calls":     func(v *WorkflowDefinitionInput) { v.Budget.ModelCalls = 21 },
		"excessive duration":  func(v *WorkflowDefinitionInput) { v.Budget.ElapsedSeconds = 181 },
		"excessive followups": func(v *WorkflowDefinitionInput) { v.Budget.FollowUpTurns = 13 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			input := workflowTestInput()
			change(&input)
			if _, err := NewWorkflowDefinition(input, workflowTestLimits()); err == nil {
				t.Fatal("invalid workflow accepted")
			}
		})
	}
}

func TestWorkflowDefinitionOwnsItsSnapshot(t *testing.T) {
	input := workflowTestInput()
	definition, err := NewWorkflowDefinition(input, workflowTestLimits())
	if err != nil {
		t.Fatal(err)
	}
	input.Instructions = "changed input"
	settings := definition.Settings()
	settings.Instructions = "changed output"
	if definition.Settings().Instructions != "" {
		t.Fatal("external mutation changed workflow snapshot")
	}
}

func TestWorkflowDefinitionRejectsInvalidOperatorLimits(t *testing.T) {
	limits := workflowTestLimits()
	for _, change := range []func(*WorkflowLimits){
		func(v *WorkflowLimits) { v.Budget.ModelCalls = 0 },
		func(v *WorkflowLimits) { v.Budget.ToolCalls = -1 },
		func(v *WorkflowLimits) { v.MaxNameRunes = 0 },
		func(v *WorkflowLimits) { v.MaxInstructionRunes = 0 },
	} {
		invalid := limits
		change(&invalid)
		if _, err := NewWorkflowDefinition(workflowTestInput(), invalid); err == nil {
			t.Fatal("invalid operator limits accepted")
		}
	}
}

func TestWorkflowDefinitionBoundsInstructionsAndProfileReferences(t *testing.T) {
	for _, change := range []func(*WorkflowDefinitionInput){
		func(v *WorkflowDefinitionInput) { v.ProviderProfileID = string([]byte{0xff}) },
		func(v *WorkflowDefinitionInput) { v.Name = strings.Repeat("家", 121) },
		func(v *WorkflowDefinitionInput) { v.Instructions = strings.Repeat("家", 4001) },
		func(v *WorkflowDefinitionInput) { v.ProviderProfileID = strings.Repeat("x", 65) },
	} {
		input := workflowTestInput()
		change(&input)
		if _, err := NewWorkflowDefinition(input, workflowTestLimits()); err == nil {
			t.Fatal("invalid workflow content accepted")
		}
	}
}

func TestWorkflowBudgetsAreIndependentExecutionCaps(t *testing.T) {
	input := workflowTestInput()
	input.Budget.ModelCalls = 2
	if _, err := NewWorkflowDefinition(input, workflowTestLimits()); err != nil {
		t.Fatal(err)
	}
	input.Budget.ToolCalls = 7
	if _, err := NewWorkflowDefinition(input, workflowTestLimits()); err == nil {
		t.Fatal("tool cap exceeded operator limit")
	}
}
