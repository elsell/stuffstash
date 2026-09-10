package agentmodel

import (
	"context"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

type completionBudgetModel struct {
	calls      int
	batch      int
	suppressed bool
}

func (m *completionBudgetModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	hasSearch := false
	for _, tool := range in.Tools {
		if tool.Name == "search" {
			hasSearch = true
		}
	}
	if !hasSearch {
		for _, message := range in.Messages {
			for _, result := range message.ToolResults {
				if result.Content == conversationDiscoveryStopped {
					m.suppressed = true
				}
			}
		}
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: fmt.Sprintf("finish-%d", m.calls), Name: "propose"}}}, nil
	}
	var calls []ports.AgentToolCall
	for i := 0; i < m.batch; i++ {
		calls = append(calls, ports.AgentToolCall{ID: fmt.Sprintf("read-%d-%d", m.calls, i), Name: "search"})
	}
	return ports.ConversationModelTurn{ToolCalls: calls}, nil
}

type completionBudgetTools struct{ reads, proposals int }

func (e *completionBudgetTools) ExecuteConversationTool(_ context.Context, call ports.AgentToolCall) (ports.ConversationToolOutcome, error) {
	if call.Name == "propose" {
		e.proposals++
		return ports.ConversationToolOutcome{ApprovalPlanID: "review"}, nil
	}
	e.reads++
	return ports.ConversationToolOutcome{Result: ports.AgentToolResult{Content: `{"assetId":"counter","parentTitle":"Kitchen"}`}}, nil
}
func TestConversationReservesCompletionWithinConfiguredLimits(t *testing.T) {
	for _, test := range []struct {
		name                        string
		models, tools, batch, reads int
		suppressed                  bool
	}{
		{"model boundary", 3, 6, 1, 2, false}, {"tool boundary", 6, 3, 4, 2, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := &completionBudgetModel{batch: test.batch}
			executor := &completionBudgetTools{}
			result, err := RunConversation(context.Background(), model, executor, ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Add coffee beans to the kitchen counter"}}, Tools: []ports.ConversationToolDefinition{{Name: "search"}, {Name: "propose"}}}, ConversationLimits{ModelCalls: test.models, ToolCalls: test.tools, FinalizationToolNames: []string{"propose"}})
			if err != nil || result.ApprovalPlanID != "review" || executor.reads != test.reads || executor.proposals != 1 || result.ModelCalls > test.models || result.ToolCalls > test.tools || model.suppressed != test.suppressed {
				t.Fatalf("completion lost: result=%+v executor=%+v model=%+v error=%v", result, executor, model, err)
			}
		})
	}
}
