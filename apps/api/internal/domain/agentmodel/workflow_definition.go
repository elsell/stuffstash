package agentmodel

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInvalidWorkflowDefinition = errors.New("invalid conversation workflow definition")

type WorkflowBudget struct {
	ToolCalls      int
	ModelCalls     int
	ElapsedSeconds int
	FollowUpTurns  int
}

// Limits are supplied by the application from operator configuration. Domain
// validation does not read environment variables or choose deployment limits.
type WorkflowLimits struct {
	Budget              WorkflowBudget
	MaxNameRunes        int
	MaxInstructionRunes int
}

type WorkflowDefinitionInput struct {
	Name              string
	ProviderProfileID string
	Instructions      string
	Budget            WorkflowBudget
}

// WorkflowDefinition owns a validated snapshot; callers can only obtain copies.
type WorkflowDefinition struct {
	settings WorkflowDefinitionInput
}

func NewWorkflowDefinition(input WorkflowDefinitionInput, limits WorkflowLimits) (WorkflowDefinition, error) {
	if !limits.valid() || !input.Budget.within(limits.Budget) {
		return WorkflowDefinition{}, ErrInvalidWorkflowDefinition
	}
	input.Name = strings.TrimSpace(input.Name)
	if !workflowTextWithin(input.Name, limits.MaxNameRunes, false) {
		return WorkflowDefinition{}, ErrInvalidWorkflowDefinition
	}
	input.Instructions = strings.TrimSpace(input.Instructions)
	input.ProviderProfileID = strings.TrimSpace(input.ProviderProfileID)
	if !workflowTextWithin(input.Instructions, limits.MaxInstructionRunes, true) || !workflowProfileReferenceValid(input.ProviderProfileID) {
		return WorkflowDefinition{}, ErrInvalidWorkflowDefinition
	}
	return WorkflowDefinition{settings: input}, nil
}

func (definition WorkflowDefinition) Settings() WorkflowDefinitionInput {
	return definition.settings
}

func (limits WorkflowLimits) valid() bool {
	return limits.Budget.positive() && limits.MaxNameRunes > 0 && limits.MaxInstructionRunes > 0
}

func (budget WorkflowBudget) positive() bool {
	return budget.ToolCalls > 0 && budget.ModelCalls > 0 && budget.ElapsedSeconds > 0 && budget.FollowUpTurns > 0
}

func (budget WorkflowBudget) within(limit WorkflowBudget) bool {
	return budget.positive() && budget.ToolCalls <= limit.ToolCalls && budget.ModelCalls <= limit.ModelCalls &&
		budget.ElapsedSeconds <= limit.ElapsedSeconds && budget.FollowUpTurns <= limit.FollowUpTurns
}

func workflowTextWithin(value string, maxRunes int, allowEmpty bool) bool {
	return utf8.ValidString(value) && (allowEmpty || value != "") && utf8.RuneCountInString(value) <= maxRunes
}

func workflowProfileReferenceValid(value string) bool {
	if value == "" {
		return true
	}
	if !utf8.ValidString(value) {
		return false
	}
	if _, ok := NewProviderProfileID(value); !ok {
		return false
	}
	return strings.IndexFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) < 0
}
