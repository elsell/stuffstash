package agentmodel

import (
	"errors"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInvalidWorkflowDefinition = errors.New("invalid conversation workflow definition")

type WorkflowStepKind string

const (
	WorkflowStepInterpret WorkflowStepKind = "interpret"
	WorkflowStepAssess    WorkflowStepKind = "assess"
	WorkflowStepRespond   WorkflowStepKind = "respond"
)

type WorkflowRetrievalStrategy string

const (
	WorkflowRetrievalPreciseFirst WorkflowRetrievalStrategy = "precise_first"
	WorkflowRetrievalExpanded     WorkflowRetrievalStrategy = "expanded"
)

type WorkflowResponseMode string

const (
	WorkflowResponseGroundedFallback WorkflowResponseMode = "generated_with_grounded_fallback"
	WorkflowResponseGrounded         WorkflowResponseMode = "grounded"
)

type WorkflowBudget struct {
	EvidenceRounds int
	ModelCalls     int
	ElapsedSeconds int
	FollowUpTurns  int
}

// Limits are supplied by the application from operator configuration. Domain
// validation does not read environment variables or choose deployment limits.
type WorkflowLimits struct {
	Budget              WorkflowBudget
	MaxStepAttempts     int
	MaxNameRunes        int
	MaxInstructionRunes int
}

type WorkflowStep struct {
	Kind              WorkflowStepKind
	ProviderProfileID string
	Instructions      string
	Attempts          int
}

type WorkflowDefinitionInput struct {
	Name      string
	Retrieval WorkflowRetrievalStrategy
	Response  WorkflowResponseMode
	Budget    WorkflowBudget
	Steps     []WorkflowStep
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
	if input.Retrieval != WorkflowRetrievalPreciseFirst && input.Retrieval != WorkflowRetrievalExpanded {
		return WorkflowDefinition{}, ErrInvalidWorkflowDefinition
	}
	if input.Response != WorkflowResponseGroundedFallback && input.Response != WorkflowResponseGrounded {
		return WorkflowDefinition{}, ErrInvalidWorkflowDefinition
	}
	order := [...]WorkflowStepKind{WorkflowStepInterpret, WorkflowStepAssess, WorkflowStepRespond}
	if len(input.Steps) != len(order) {
		return WorkflowDefinition{}, ErrInvalidWorkflowDefinition
	}
	input.Steps = slices.Clone(input.Steps)
	for index := range input.Steps {
		step := &input.Steps[index]
		step.Instructions = strings.TrimSpace(step.Instructions)
		step.ProviderProfileID = strings.TrimSpace(step.ProviderProfileID)
		if step.Kind != order[index] || step.Attempts < 1 || step.Attempts > limits.MaxStepAttempts ||
			!workflowTextWithin(step.Instructions, limits.MaxInstructionRunes, true) || !workflowProfileReferenceValid(step.ProviderProfileID) {
			return WorkflowDefinition{}, ErrInvalidWorkflowDefinition
		}
	}
	return WorkflowDefinition{settings: input}, nil
}

func (definition WorkflowDefinition) Settings() WorkflowDefinitionInput {
	settings := definition.settings
	settings.Steps = slices.Clone(settings.Steps)
	return settings
}

func (limits WorkflowLimits) valid() bool {
	return limits.Budget.positive() && limits.MaxStepAttempts > 0 && limits.MaxNameRunes > 0 && limits.MaxInstructionRunes > 0
}

func (budget WorkflowBudget) positive() bool {
	return budget.EvidenceRounds > 0 && budget.ModelCalls > 0 && budget.ElapsedSeconds > 0 && budget.FollowUpTurns > 0
}

func (budget WorkflowBudget) within(limit WorkflowBudget) bool {
	return budget.positive() && budget.EvidenceRounds <= limit.EvidenceRounds && budget.ModelCalls <= limit.ModelCalls &&
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
