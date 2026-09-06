package httpserver

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// These controlled providers choose fixed fixture actions. They exercise the
// real authorized reads and proposal boundary without implementing an intent engine.
func (actionPlanProposalLanguageModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return httpConversationProposal(in, "create_asset", "water bottle", "")
}
func (moveActionPlanProposalLanguageModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return httpConversationProposal(in, "move_asset", "water bottle", "Office")
}
func (archiveActionPlanProposalLanguageModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	subject := "water bottle"
	for _, message := range in.Messages {
		if message.Role == ports.ConversationRoleUser && strings.Contains(strings.ToLower(message.Text), "toolbox") {
			subject = "Toolbox"
		}
	}
	return httpConversationProposal(in, "archive_asset", subject, "")
}
func (restoreActionPlanProposalLanguageModel) Converse(_ context.Context, in ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return httpConversationProposal(in, "restore_asset", "water bottle", "")
}

func httpConversationProposal(in ports.ConversationModelInput, kind, subject, destination string) (ports.ConversationModelTurn, error) {
	if len(in.Messages) == 0 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	if in.Messages[len(in.Messages)-1].Role == ports.ConversationRoleUser {
		lifecycle := "active"
		if kind == "restore_asset" {
			lifecycle = "archived"
		}
		calls := []ports.AgentToolCall{{ID: "find-subject", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": subject, "lifecycleState": lifecycle}}}
		if destination != "" {
			calls = append(calls, ports.AgentToolCall{ID: "find-parent", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": destination}})
		}
		return ports.ConversationModelTurn{ToolCalls: calls}, nil
	}
	expectedReads := []string{"find-subject"}
	if destination != "" {
		expectedReads = append(expectedReads, "find-parent")
	}
	ids, err := httpConversationEvidenceIDs(in, expectedReads...)
	if err != nil {
		return ports.ConversationModelTurn{}, err
	}
	arguments := map[string]any{}
	verb := "Create"
	id := "create-subject"
	if kind == "create_asset" {
		arguments["title"], arguments["kind"] = subject, "item"
	} else {
		assetID := ids[strings.ToLower(subject)]
		if assetID == "" {
			return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
		}
		arguments["assetId"] = assetID
		switch kind {
		case "move_asset":
			verb, id = "Move", "move-subject"
		case "archive_asset":
			verb, id = "Archive", "archive-subject"
		case "restore_asset":
			verb, id = "Restore", "restore-subject"
		default:
			return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
		}
	}
	summary := verb + " " + subject
	if destination != "" {
		parentID := ids[strings.ToLower(destination)]
		if parentID == "" {
			return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
		}
		arguments["parentAssetId"] = parentID
		summary += " to " + destination
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "propose-change", Name: "propose_inventory_change", Arguments: map[string]any{
		"summary": summary + "?", "commands": []any{map[string]any{"id": id, "kind": kind, "summary": summary, "arguments": arguments}},
	}}}}, nil
}

func httpConversationEvidenceIDs(in ports.ConversationModelInput, expectedReads ...string) (map[string]string, error) {
	ids := map[string]string{}
	completed := map[string]bool{}
	for _, message := range in.Messages {
		for _, result := range message.ToolResults {
			var output struct {
				Tool  string          `json:"tool"`
				Error json.RawMessage `json:"error"`
				Items []struct {
					AssetID string `json:"assetId"`
					Title   string `json:"title"`
				} `json:"items"`
			}
			if json.Unmarshal([]byte(result.Content), &output) != nil || len(output.Error) != 0 || output.Tool != app.RealtimeVoiceToolSearchAuthorizedAssets {
				return nil, ports.ErrInvalidProviderInput
			}
			completed[result.CallID] = true
			for _, item := range output.Items {
				ids[strings.ToLower(item.Title)] = item.AssetID
			}
		}
	}
	for _, id := range expectedReads {
		if !completed[id] {
			return nil, ports.ErrInvalidProviderInput
		}
	}
	return ids, nil
}
