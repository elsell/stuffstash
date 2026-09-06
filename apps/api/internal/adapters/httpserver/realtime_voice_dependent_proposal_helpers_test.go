package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (createNestedItemActionPlanProposalLanguageModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if len(in.Messages) == 0 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	if in.Messages[len(in.Messages)-1].Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{
			{ID: "find-refills", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "diaper genie refills"}},
			{ID: "find-room", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Henry's room"}},
			{ID: "find-closet", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "closet"}},
		}}, nil
	}
	ids, err := httpConversationEvidenceIDs(in, "find-refills", "find-room", "find-closet")
	if err != nil || len(ids) != 0 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "propose-nested-create", Name: "propose_inventory_change", Arguments: map[string]any{
		"summary": "Add diaper genie refills to the closet in Henry's room?",
		"commands": []any{
			map[string]any{"id": "create-destination-0", "kind": "create_location", "summary": "Create Henry's room", "arguments": map[string]any{"title": "Henry's room"}},
			map[string]any{"id": "create-destination-1", "kind": "create_asset", "summary": "Create closet", "arguments": map[string]any{"title": "closet", "kind": "container", "parentCommandId": "create-destination-0"}},
			map[string]any{"id": "create-subject", "kind": "create_asset", "summary": "Create diaper genie refills", "arguments": map[string]any{"title": "diaper genie refills", "kind": "item", "parentCommandId": "create-destination-1"}},
		},
	}}}}, nil
}

func (moveToMissingLocationActionPlanProposalLanguageModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if len(in.Messages) == 0 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	if in.Messages[len(in.Messages)-1].Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{
			{ID: "find-bottle", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "water bottle"}},
			{ID: "find-kitchen", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Kitchen"}},
		}}, nil
	}
	ids, err := httpConversationEvidenceIDs(in, "find-bottle", "find-kitchen")
	if err != nil || ids["water bottle"] == "" || ids["kitchen"] != "" {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "propose-dependent-move", Name: "propose_inventory_change", Arguments: map[string]any{
		"summary": "Move Water bottle to Kitchen?",
		"commands": []any{
			map[string]any{"id": "create-destination-0", "kind": "create_location", "summary": "Create Kitchen", "arguments": map[string]any{"title": "Kitchen"}},
			map[string]any{"id": "move-subject", "kind": "move_asset", "summary": "Move Water bottle to Kitchen", "arguments": map[string]any{"assetId": ids["water bottle"], "parentCommandId": "create-destination-0"}},
		},
	}}}}, nil
}
