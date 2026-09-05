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
		return ports.ConversationModelTurn{ProviderState: []byte("opaque-provider-state"), ToolCalls: []ports.AgentToolCall{{ID: "search-1", Name: "search", Arguments: map[string]any{"query": "chemicals"}}}}, nil
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
	if len(result.Messages) < 2 || string(result.Messages[1].ProviderState) != "opaque-provider-state" {
		t.Fatal("provider continuation was lost")
	}
	for _, message := range result.Messages {
		if message.Role == ports.ConversationRoleTool && len(message.ToolResults) == 1 && message.ToolResults[0].CallID == "search-1" {
			found = true
		}
	}
	if !found {
		t.Fatal("follow-up history lost the executed search result")
	}
}

type conversationProposalModel struct{ calls int }

func (m *conversationProposalModel) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{
		{ID: "proposal-1", Name: "propose_change", Arguments: map[string]any{"assetId": "existing-drill"}},
		{ID: "later-read", Name: "search", Arguments: map[string]any{"query": "garage"}},
	}}, nil
}

type conversationProposalTools struct{ calls int }

func (e *conversationProposalTools) ExecuteConversationTool(_ context.Context, call ports.AgentToolCall) (ports.ConversationToolOutcome, error) {
	e.calls++
	return ports.ConversationToolOutcome{Result: ports.AgentToolResult{CallID: call.ID, Name: call.Name, Content: "Change proposed for review."}, ApprovalPlanID: "review-plan"}, nil
}
func TestConversationLoopPausesImmediatelyForApproval(t *testing.T) {
	model, executor := &conversationProposalModel{}, &conversationProposalTools{}
	result, err := RunConversation(context.Background(), model, executor, ports.ConversationModelInput{
		Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Move my drill into the garage."}},
	}, ConversationLimits{ModelCalls: 4, ToolCalls: 4})
	if err != nil {
		t.Fatal(err)
	}
	if result.ApprovalPlanID != "review-plan" || result.Answer != nil || model.calls != 1 || executor.calls != 1 {
		t.Fatalf("continued beyond approval boundary: result=%+v model=%d tools=%d", result, model.calls, executor.calls)
	}
}
func TestConversationLoopCancellationDoesNotCallProviders(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	model, executor := &conversationInventoryModel{}, &conversationInventoryTools{}
	_, err := RunConversation(ctx, model, executor, ports.ConversationModelInput{
		Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Do I have chemicals?"}},
	}, ConversationLimits{ModelCalls: 3, ToolCalls: 3})
	if err != context.Canceled || model.calls != 0 || executor.calls != 0 {
		t.Fatalf("cancelled conversation ran work: error=%v model=%d tools=%d", err, model.calls, executor.calls)
	}
}
func TestConversationLoopModelBudgetRetainsExecutedEvidence(t *testing.T) {
	model, executor := &conversationInventoryModel{}, &conversationInventoryTools{}
	result, err := RunConversation(context.Background(), model, executor, ports.ConversationModelInput{
		Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Do I have chemicals?"}},
	}, ConversationLimits{ModelCalls: 1, ToolCalls: 3})
	if err == nil || model.calls != 1 || executor.calls != 1 {
		t.Fatalf("budget not enforced: %v %d %d", err, model.calls, executor.calls)
	}
	if len(result.Messages) == 0 || len(result.Messages[len(result.Messages)-1].ToolResults) != 1 {
		t.Fatal("budget exhaustion discarded useful evidence")
	}
}

func TestConversationApprovalHistoryClosesUnexecutedCalls(t *testing.T) {
	result, err := RunConversation(context.Background(), &conversationProposalModel{}, &conversationProposalTools{}, ports.ConversationModelInput{
		Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Move my drill."}},
	}, ConversationLimits{ModelCalls: 4, ToolCalls: 4})
	if err != nil {
		t.Fatal(err)
	}
	answered := map[string]bool{}
	for _, message := range result.Messages {
		for _, tool := range message.ToolResults {
			answered[tool.CallID] = true
		}
	}
	if !answered["proposal-1"] || !answered["later-read"] {
		t.Fatal("approval left unresolved native tool calls in history")
	}
}

type conversationCancellingTools struct{ cancel context.CancelFunc }

func (e conversationCancellingTools) ExecuteConversationTool(_ context.Context, call ports.AgentToolCall) (ports.ConversationToolOutcome, error) {
	e.cancel()
	return ports.ConversationToolOutcome{Result: ports.AgentToolResult{CallID: call.ID, Name: call.Name, Content: "Proposed, not executed."}, ApprovalPlanID: "review-plan"}, nil
}
func TestConversationCancellationDuringProposalPreservesOutcomeWithoutSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := RunConversation(ctx, &conversationProposalModel{}, conversationCancellingTools{cancel: cancel}, ports.ConversationModelInput{
		Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Move my drill."}},
	}, ConversationLimits{ModelCalls: 4, ToolCalls: 4})
	if err != context.Canceled || result.ApprovalPlanID != "review-plan" {
		t.Fatalf("lost cancellation or durable proposal: %+v %v", result, err)
	}
}
