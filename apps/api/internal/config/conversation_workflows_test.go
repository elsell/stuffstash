package config

import (
	"strings"
	"testing"
)

func TestWorkflowLimitsCaptureEnvironmentAndRejectInvalidValues(t *testing.T) {
	names := []string{"STUFF_STASH_WORKFLOW_MAX_TOOL_CALLS", "STUFF_STASH_WORKFLOW_MAX_MODEL_CALLS", "STUFF_STASH_WORKFLOW_MAX_ELAPSED_SECONDS", "STUFF_STASH_WORKFLOW_MAX_FOLLOW_UP_TURNS", "STUFF_STASH_WORKFLOW_MAX_NAME_RUNES", "STUFF_STASH_WORKFLOW_MAX_INSTRUCTION_RUNES"}
	for _, name := range names {
		t.Setenv(name, "")
	}
	defaults, err := Load().ConversationWorkflows.Limits()
	if err != nil || defaults.Budget.ToolCalls != 12 || defaults.Budget.ModelCalls != 12 || defaults.MaxInstructionRunes != 4000 {
		t.Fatalf("defaults: %+v %v", defaults, err)
	}
	t.Setenv(names[0], "7")
	captured := Load()
	t.Setenv(names[0], "9")
	limits, err := captured.ConversationWorkflows.Limits()
	if err != nil || limits.Budget.ToolCalls != 7 {
		t.Fatalf("configuration was not captured: %+v %v", limits, err)
	}
	t.Setenv(names[0], "")
	for _, name := range names {
		for _, value := range []string{"0", "-1", "not-an-integer", "99999999999999999999999999"} {
			t.Run(name+"/"+value, func(t *testing.T) {
				t.Setenv(name, value)
				_, err := Load().ConversationWorkflows.Limits()
				if err == nil || !strings.Contains(err.Error(), name) || strings.Contains(err.Error(), value) {
					t.Fatalf("expected safe named configuration error, got %v", err)
				}
			})
		}
	}
}
func TestProgrammaticWorkflowConfigurationUsesOperatorDefaults(t *testing.T) {
	var cfg Config
	limits, err := cfg.ConversationWorkflows.Limits()
	if err != nil || limits.Budget.ElapsedSeconds != 60 || limits.Budget.ToolCalls != 12 {
		t.Fatalf("zero configuration: %+v %v", limits, err)
	}
}

func TestWorkflowRejectsRetiredOperatorSettings(t *testing.T) {
	for _, name := range []string{"STUFF_STASH_WORKFLOW_MAX_EVIDENCE_ROUNDS", "STUFF_STASH_WORKFLOW_MAX_STEP_ATTEMPTS"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, "2")
			if _, err := Load().ConversationWorkflows.Limits(); err == nil || !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), "remove") {
				t.Fatalf("missing migration instruction: %v", err)
			}
		})
	}
}

func TestConversationContextLimitCapturesOperatorSetting(t *testing.T) {
	t.Setenv("STUFF_STASH_CONVERSATION_MAX_CONTEXT_BYTES", "4096")
	captured := Load()
	t.Setenv("STUFF_STASH_CONVERSATION_MAX_CONTEXT_BYTES", "8192")
	limit, err := captured.ConversationWorkflows.ContextBytes()
	if err != nil || limit != 4096 {
		t.Fatalf("context setting was not captured: limit=%d err=%v", limit, err)
	}
	for _, invalid := range []string{"0", "-1", "unbounded"} {
		t.Setenv("STUFF_STASH_CONVERSATION_MAX_CONTEXT_BYTES", invalid)
		if _, err := Load().ConversationWorkflows.ContextBytes(); err == nil {
			t.Fatalf("accepted invalid context cap %q", invalid)
		}
	}
}
