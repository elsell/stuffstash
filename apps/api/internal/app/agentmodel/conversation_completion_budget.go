package agentmodel

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const conversationDiscoveryStopped = "This discovery call was not executed: the remaining budget is reserved for completing the request. Use the evidence already returned; do not infer a result for this call."

func isConversationFinalizationTool(limits ConversationLimits, name string) bool {
	for _, terminal := range limits.FinalizationToolNames {
		if name == terminal {
			return true
		}
	}
	return false
}
func reserveConversationCompletion(limits ConversationLimits, result ConversationResult) bool {
	return len(limits.FinalizationToolNames) > 0 && (result.ModelCalls >= limits.ModelCalls || result.ToolCalls >= limits.ToolCalls-1)
}
func conversationBudgetInput(input ports.ConversationModelInput, limits ConversationLimits, result ConversationResult) ports.ConversationModelInput {
	if len(limits.FinalizationToolNames) == 0 {
		return input
	}
	input.Instructions += fmt.Sprintf("\nRequest budget including this response: %d model responses and %d tool calls remain. Reserve the final response and tool call for a proposal or answer. Stop discovery as soon as the evidence supports completing the request.", limits.ModelCalls-result.ModelCalls+1, limits.ToolCalls-result.ToolCalls)
	if !reserveConversationCompletion(limits, result) {
		return input
	}
	tools := make([]ports.ConversationToolDefinition, 0, len(input.Tools))
	for _, tool := range input.Tools {
		if isConversationFinalizationTool(limits, tool.Name) {
			tools = append(tools, tool)
		}
	}
	input.Tools = tools
	input.Instructions += "\nDiscovery is now closed. Complete the request using the evidence already returned: propose supported changes for approval or deliver an evidence-based answer. If a specific unresolved ambiguity prevents a safe proposal, ask that concrete question. Do not claim a change was executed, invent IDs, or tell the person to simplify the request because of the budget."
	return input
}
