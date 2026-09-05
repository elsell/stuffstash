package agentmodel

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

var ErrConversationBudgetExhausted = errors.New("conversation call budget exhausted")

type ConversationLimits struct{ ModelCalls, ToolCalls int }
type ConversationResult struct {
	Answer         *ports.ConversationAnswer
	Messages       []ports.ConversationMessage
	ApprovalPlanID string
	ModelCalls     int
	ToolCalls      int
}

// RunConversation coordinates model turns and scoped tools. The caller owns
// the overall deadline and the executor owns authorization and tool validation.
func RunConversation(ctx context.Context, model ports.ConversationModel, executor ports.ConversationToolExecutor, input ports.ConversationModelInput, limits ConversationLimits) (result ConversationResult, err error) {
	result = ConversationResult{Messages: append([]ports.ConversationMessage(nil), input.Messages...)}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if model == nil || executor == nil || len(input.Messages) == 0 || limits.ModelCalls < 1 || limits.ToolCalls < 0 {
		return result, ports.ErrInvalidProviderInput
	}
	defer closeUnexecutedConversationCalls(&result, len(input.Messages))
	seenCalls := map[string]struct{}{}
	for result.ModelCalls < limits.ModelCalls {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		input.Messages = result.Messages
		result.ModelCalls++
		turn, err := model.Converse(ctx, input)
		if err != nil {
			return result, err
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if len(turn.ToolCalls) == 0 {
			answer := turn.Answer
			if answer == nil && strings.TrimSpace(turn.Text) != "" {
				answer = &ports.ConversationAnswer{Spoken: turn.Text, Display: turn.Text}
			}
			if answer == nil || strings.TrimSpace(answer.Spoken) == "" || strings.TrimSpace(answer.Display) == "" {
				return result, ports.ErrInvalidProviderInput
			}
			result.Answer = answer
			result.Messages = append(result.Messages, ports.ConversationMessage{Role: ports.ConversationRoleAssistant, Text: answer.Display, ProviderState: turn.ProviderState})
			return result, nil
		}
		if turn.Answer != nil {
			return result, ports.ErrInvalidProviderInput
		}
		// Validate call identities before executing any member of this model turn.
		for _, call := range turn.ToolCalls {
			if strings.TrimSpace(call.ID) == "" || strings.TrimSpace(call.Name) == "" {
				return result, ports.ErrInvalidProviderInput
			}
			if _, duplicate := seenCalls[call.ID]; duplicate {
				return result, ports.ErrInvalidProviderInput
			}
			seenCalls[call.ID] = struct{}{}
		}
		result.Messages = append(result.Messages, ports.ConversationMessage{Role: ports.ConversationRoleAssistant, Text: turn.Text, ToolCalls: turn.ToolCalls, ProviderState: turn.ProviderState})
		for _, call := range turn.ToolCalls {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			if result.ToolCalls >= limits.ToolCalls {
				return result, ErrConversationBudgetExhausted
			}
			result.ToolCalls++
			outcome, err := executor.ExecuteConversationTool(ctx, call)
			if err != nil {
				return result, err
			}
			// Correlation is application-owned, even if an executor omitted metadata.
			outcome.Result.CallID, outcome.Result.Name, outcome.Result.Call = call.ID, call.Name, call
			result.Messages = append(result.Messages, ports.ConversationMessage{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{outcome.Result}})
			result.ApprovalPlanID = outcome.ApprovalPlanID
			if err := ctx.Err(); err != nil {
				return result, err
			}
			if result.ApprovalPlanID != "" {
				return result, nil
			}
		}
	}
	return result, ErrConversationBudgetExhausted
}

// A paused or failed turn must not leave native tool-call messages unanswered.
// These results explicitly describe non-execution; they never run skipped work.
func closeUnexecutedConversationCalls(result *ConversationResult, start int) {
	completed := map[string]bool{}
	var requested []ports.AgentToolCall
	for _, message := range result.Messages[start:] {
		requested = append(requested, message.ToolCalls...)
		for _, tool := range message.ToolResults {
			completed[tool.CallID] = true
		}
	}
	for _, call := range requested {
		if completed[call.ID] {
			continue
		}
		result.Messages = append(result.Messages, ports.ConversationMessage{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{{CallID: call.ID, Name: call.Name, Call: call, Content: "This tool call was not completed because the conversation stopped. Do not infer a result or repeat a change without checking its status."}}})
	}
}
