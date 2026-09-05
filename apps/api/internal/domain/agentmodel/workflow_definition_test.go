package agentmodel

import "testing"

func workflowTestInput() WorkflowDefinitionInput {
	return WorkflowDefinitionInput{
		Name: "Household voice", Retrieval: WorkflowRetrievalPreciseFirst, Response: WorkflowResponseGroundedFallback,
		Budget: WorkflowBudget{EvidenceRounds: 3, ModelCalls: 8, ElapsedSeconds: 60, FollowUpTurns: 6},
		Steps: []WorkflowStep{
			{Kind: WorkflowStepInterpret, Attempts: 2},
			{Kind: WorkflowStepAssess, Attempts: 2},
			{Kind: WorkflowStepRespond, Attempts: 1},
		},
	}
}

func workflowTestLimits() WorkflowLimits {
	return WorkflowLimits{Budget: WorkflowBudget{EvidenceRounds: 6, ModelCalls: 20, ElapsedSeconds: 180, FollowUpTurns: 12}, MaxStepAttempts: 3, MaxNameRunes: 120, MaxInstructionRunes: 4000}
}

func TestWorkflowDefinitionAcceptsTuningWithinOperatorLimits(t *testing.T) {
	input := workflowTestInput()
	input.Steps[0].Instructions = "Prefer household tags when names differ."
	definition, err := NewWorkflowDefinition(input, workflowTestLimits())
	if err != nil {
		t.Fatal(err)
	}
	if definition.Settings().Steps[0].Instructions != input.Steps[0].Instructions || definition.Settings().Budget != input.Budget {
		t.Fatal("workflow tuning was discarded")
	}
}

func TestWorkflowDefinitionRejectsInvalidPolicy(t *testing.T) {
	cases := map[string]func(*WorkflowDefinitionInput){
		"empty name":           func(v *WorkflowDefinitionInput) { v.Name = " " },
		"zero evidence budget": func(v *WorkflowDefinitionInput) { v.Budget.EvidenceRounds = 0 },
		"excessive calls":      func(v *WorkflowDefinitionInput) { v.Budget.ModelCalls = 21 },
		"excessive duration":   func(v *WorkflowDefinitionInput) { v.Budget.ElapsedSeconds = 181 },
		"excessive followups":  func(v *WorkflowDefinitionInput) { v.Budget.FollowUpTurns = 13 },
		"unknown retrieval":    func(v *WorkflowDefinitionInput) { v.Retrieval = "unrestricted" },
		"unknown response":     func(v *WorkflowDefinitionInput) { v.Response = "execute_model_text" },
		"missing step":         func(v *WorkflowDefinitionInput) { v.Steps = v.Steps[:2] },
		"duplicate step":       func(v *WorkflowDefinitionInput) { v.Steps[1].Kind = WorkflowStepInterpret },
		"unknown step":         func(v *WorkflowDefinitionInput) { v.Steps[0].Kind = "bypass_authorization" },
		"zero attempts":        func(v *WorkflowDefinitionInput) { v.Steps[0].Attempts = 0 },
		"excessive attempts":   func(v *WorkflowDefinitionInput) { v.Steps[0].Attempts = 4 },
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
	input.Steps[0].Instructions = "changed input"
	settings := definition.Settings()
	settings.Steps[1].Instructions = "changed output"
	if definition.Settings().Steps[0].Instructions != "" || definition.Settings().Steps[1].Instructions != "" {
		t.Fatal("external mutation changed workflow snapshot")
	}
}
