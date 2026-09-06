package voice

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestDevelopmentConversationUsesReturnedIdentityAndParent(t *testing.T) {
	model, ok := any(DevFakeLanguageInference{}).(ports.ConversationModel)
	if !ok {
		t.Fatal("development model does not implement conversation port")
	}
	input := ports.ConversationModelInput{Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Where are my tools?"}}}
	turn, err := model.Converse(context.Background(), input)
	if err != nil || len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "search_authorized_assets" {
		t.Fatalf("development demo did not request authorized search: %+v %v", turn, err)
	}
	call := turn.ToolCalls[0]
	input.Messages = append(input.Messages, ports.ConversationMessage{Role: ports.ConversationRoleAssistant, ToolCalls: turn.ToolCalls}, ports.ConversationMessage{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{{CallID: call.ID, Name: call.Name, Content: `{"tool":"search_authorized_assets","items":[{"assetId":"observed-toolbox","title":"Tools","parentTitle":"Garage"}]}`}}})
	answer, err := model.Converse(context.Background(), input)
	if err != nil || answer.Answer == nil || answer.Answer.Spoken != "I found Tools in Garage." || len(answer.Answer.AssetIDs) != 1 || answer.Answer.AssetIDs[0] != "observed-toolbox" {
		t.Fatalf("development demo lost returned identity/location: %+v %v", answer, err)
	}
}
