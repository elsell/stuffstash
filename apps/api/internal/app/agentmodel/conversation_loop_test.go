package agentmodel

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

// This fake changes its answer from actual tool feedback; it does not assert an
// interpretation phase, semantic operation or wording template.
type conversationInventoryModel struct{ calls int }

func (m *conversationInventoryModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	last := in.Messages[len(in.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "search-1", Name: "search", Arguments: map[string]any{"query": "chemicals"}}}}, nil
	}
	var titles []string
	if len(last.ToolResults) > 0 {
		_ = json.Unmarshal([]byte(last.ToolResults[0].Content), &titles)
	}
	if len(titles) == 0 {
		return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: "I couldn't find any chemicals in the results.", Display: "No matching items found."}}, nil
	}
	return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: "Yes, you have chemicals, including " + titles[0] + ".", Display: "Matching chemicals: " + titles[0]}}, nil
}

type conversationInventoryTools struct{ calls int }

func (e *conversationInventoryTools) ExecuteConversationTool(_ context.Context, call ports.AgentToolCall) (ports.ConversationToolOutcome, error) {
	e.calls++
	return ports.ConversationToolOutcome{Result: ports.AgentToolResult{CallID: call.ID, Name: call.Name, Call: call, Content: `["Isopropyl alcohol","Acetone"]`}}, nil
}

func TestConversationLoopSearchesAndPreservesNaturalAnswer(t *testing.T) {
	model, executor := &conversationInventoryModel{}, &conversationInventoryTools{}
	result, err := RunConversation(context.Background(), model, executor, ports.ConversationModelInput{
		Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Do I have any chemicals?"}},
	}, ConversationLimits{ModelCalls: 3, ToolCalls: 3})
	if err != nil {
		t.Fatal(err)
	}
	if result.Answer == nil || result.Answer.Spoken != "Yes, you have chemicals, including Isopropyl alcohol." {
		t.Fatalf("model answer lost: %+v", result)
	}
	if model.calls != 2 || executor.calls != 1 {
		t.Fatalf("redundant stages: model=%d tools=%d", model.calls, executor.calls)
	}
}

func TestConversationLoopRetainsToolResultsInHistory(t *testing.T) {
	result, err := RunConversation(context.Background(), &conversationInventoryModel{}, &conversationInventoryTools{}, ports.ConversationModelInput{
		Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Do I have any chemicals?"}},
	}, ConversationLimits{ModelCalls: 3, ToolCalls: 3})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, message := range result.Messages {
		if message.Role == ports.ConversationRoleTool && len(message.ToolResults) == 1 && message.ToolResults[0].CallID == "search-1" {
			found = true
		}
	}
	if !found {
		t.Fatal("follow-up history lost the executed search result")
	}
}
