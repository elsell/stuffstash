package config

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"os"
	"strconv"
	"strings"
)

// WorkflowConfiguration captures environment-backed limits before application construction.
// Its zero value provides the documented defaults for programmatic callers.
type WorkflowConfiguration struct {
	contextBytes int
	limits       agentmodel.WorkflowLimits
	captured     bool
	err          error
}

func defaultWorkflowLimits() agentmodel.WorkflowLimits {
	return agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{EvidenceRounds: 4, ModelCalls: 12, ElapsedSeconds: 60, FollowUpTurns: 8}, MaxStepAttempts: 2, MaxNameRunes: 100, MaxInstructionRunes: 4000}
}
func (c WorkflowConfiguration) Limits() (agentmodel.WorkflowLimits, error) {
	if c.err != nil {
		return agentmodel.WorkflowLimits{}, c.err
	}
	if !c.captured {
		return defaultWorkflowLimits(), nil
	}
	return c.limits, nil
}
func loadWorkflowConfiguration() WorkflowConfiguration {
	result := WorkflowConfiguration{limits: defaultWorkflowLimits(), contextBytes: 2 * 1024 * 1024, captured: true}
	entries := []struct {
		name   string
		target *int
	}{
		{"STUFF_STASH_CONVERSATION_MAX_CONTEXT_BYTES", &result.contextBytes},
		{"STUFF_STASH_WORKFLOW_MAX_EVIDENCE_ROUNDS", &result.limits.Budget.EvidenceRounds},
		{"STUFF_STASH_WORKFLOW_MAX_MODEL_CALLS", &result.limits.Budget.ModelCalls},
		{"STUFF_STASH_WORKFLOW_MAX_ELAPSED_SECONDS", &result.limits.Budget.ElapsedSeconds},
		{"STUFF_STASH_WORKFLOW_MAX_FOLLOW_UP_TURNS", &result.limits.Budget.FollowUpTurns},
		{"STUFF_STASH_WORKFLOW_MAX_STEP_ATTEMPTS", &result.limits.MaxStepAttempts},
		{"STUFF_STASH_WORKFLOW_MAX_NAME_RUNES", &result.limits.MaxNameRunes},
		{"STUFF_STASH_WORKFLOW_MAX_INSTRUCTION_RUNES", &result.limits.MaxInstructionRunes},
	}
	for _, entry := range entries {
		raw := strings.TrimSpace(os.Getenv(entry.name))
		if raw == "" {
			continue
		}
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			result.err = fmt.Errorf("%s must be a positive integer", entry.name)
			return result
		}
		*entry.target = value
	}
	if result.limits.Budget.EvidenceRounds > agentmodel.MaxEvidenceRounds {
		result.err = fmt.Errorf("STUFF_STASH_WORKFLOW_MAX_EVIDENCE_ROUNDS exceeds the supported investigation ceiling")
	}
	return result
}

func (c WorkflowConfiguration) ContextBytes() (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	if !c.captured {
		return 2 * 1024 * 1024, nil
	}
	return c.contextBytes, nil
}
