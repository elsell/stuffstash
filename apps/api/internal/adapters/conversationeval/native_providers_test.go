package conversationeval

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type nativeEvaluationModel struct {
	taggedClothesModel
	err error
}

type nativeFixtureModel struct{ taggedClothesModel }

func (*nativeFixtureModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	last := input.Messages[len(input.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "find", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "baby clothes"}}}}, nil
	}
	var evidence struct {
		Items []struct {
			AssetID     string `json:"assetId"`
			ParentTitle string `json:"parentTitle"`
		} `json:"items"`
	}
	if len(last.ToolResults) != 1 || json.Unmarshal([]byte(last.ToolResults[0].Content), &evidence) != nil || len(evidence.Items) != 1 {
		return ports.ConversationModelTurn{}, errors.New("missing fixture evidence")
	}
	item := evidence.Items[0]
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "answer", Name: "present_answer", Arguments: map[string]any{"spoken": "They are in " + item.ParentTitle + ".", "display": "Here are your matching clothes.", "assetIds": []string{item.AssetID}}}}}, nil
}

func TestNativeTextEvaluationUsesProductionSearchAndPresentation(t *testing.T) {
	model := &nativeFixtureModel{}
	input := testInput(t, model)
	input.Providers.ConversationModel = model
	result, err := New(Dependencies{Clock: testClock{}, IDs: &testIDs{}}).Execute(context.Background(), input)
	if err != nil || !input.Case.Evaluate(result.Outcome).Passed || result.ModelCalls != 2 || result.Coverage != ports.EvaluationCoverageText {
		t.Fatalf("native evaluation differs from production conversation: %+v %v", result, err)
	}
}

func (m *nativeEvaluationModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if input.TenantID != "tenant-home" || input.Instructions != "tenant guidance" || len(input.Messages) != 1 || input.Messages[0].Text != "Where are my clothes?" {
		return ports.ConversationModelTurn{}, errors.New("conversation context changed")
	}
	return ports.ConversationModelTurn{Text: "They are in the hall closet."}, m.err
}

type explicitNativeEvaluationProvider struct{ model *nativeEvaluationModel }

func (p explicitNativeEvaluationProvider) ResolveWorkflowLanguageProvider(_ context.Context, input ports.WorkflowLanguageProviderResolutionInput) (ports.WorkflowLanguageProviderBinding, error) {
	return ports.WorkflowLanguageProviderBinding{ProfileID: input.ProfileID, PromptTemplate: "tenant guidance", Provider: p.model}, nil
}

func TestEvaluationPreservesAndCountsNativeProviderCalls(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		for _, failed := range []bool{false, true} {
			name := "default"
			if explicit {
				name = "explicit"
			}
			if failed {
				name += " failure"
			}
			t.Run(name, func(t *testing.T) {
				model := &nativeEvaluationModel{}
				if failed {
					model.err = errors.New("provider unavailable")
				}
				calls := &atomic.Int64{}
				resolver := textProviders{providers: ports.RealtimeVoiceProviderSet{ConversationModel: model}, explicit: explicitNativeEvaluationProvider{model}, calls: calls}
				var conversation ports.ConversationModel
				if explicit {
					binding, err := resolver.ResolveWorkflowLanguageProvider(context.Background(), ports.WorkflowLanguageProviderResolutionInput{TenantID: "tenant-home", ProfileID: "selected"})
					if err != nil {
						t.Fatal(err)
					}
					conversation = binding.Provider
				} else {
					providers, err := resolver.ResolveRealtimeVoiceProviders(context.Background(), ports.RealtimeVoiceProviderResolutionInput{})
					if err != nil {
						t.Fatal(err)
					}
					conversation = providers.ConversationModel
				}
				if conversation == nil {
					t.Fatal("native provider hidden by evaluation wrapper")
				}
				turn, err := conversation.Converse(context.Background(), ports.ConversationModelInput{TenantID: "tenant-home", Instructions: "tenant guidance", Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Where are my clothes?"}}})
				if !errors.Is(err, model.err) || turn.Text != "They are in the hall closet." || calls.Load() != 1 {
					t.Fatalf("native context, result or call count lost: turn=%+v err=%v calls=%d", turn, err, calls.Load())
				}
			})
		}
	}
}
