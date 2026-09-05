package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

const googleConversationMaxResponseBytes = 1024 * 1024
const googleConversationMaxStateBytes = 256 * 1024

// Native conversation is independent of the obsolete investigation schema.
// One Converse call makes one provider request so application budgets stay real.
func (p GoogleGeminiLanguageInference) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	request, err := googleConversationRequest(input)
	if err != nil {
		return ports.ConversationModelTurn{}, err
	}
	var response googleConversationResponse
	client := p.client
	client.maxResponseBytes = googleConversationMaxResponseBytes
	if err := client.postJSON(ctx, p.path, request, &response); err != nil {
		return ports.ConversationModelTurn{}, err
	}
	if len(response.Candidates) == 0 || response.Candidates[0].FinishReason != "STOP" {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	content := response.Candidates[0].Content
	state, err := json.Marshal(content)
	if err != nil {
		return ports.ConversationModelTurn{}, err
	}
	if len(state) > googleConversationMaxStateBytes {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	turn := ports.ConversationModelTurn{ProviderState: state}
	var text []string
	for index, raw := range content.Parts {
		var part googleConversationPart
		if json.Unmarshal(raw, &part) != nil {
			return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
		}
		if !part.Thought && part.Text != "" {
			text = append(text, part.Text)
		}
		if part.FunctionCall != nil {
			call := part.FunctionCall
			id := call.ID
			if id == "" {
				id = fmt.Sprintf("turn-%d-call-%d", len(input.Messages), index)
			}
			turn.ToolCalls = append(turn.ToolCalls, ports.AgentToolCall{ID: id, Name: call.Name, Arguments: call.Args})
		}
	}
	turn.Text = strings.Join(text, "\n")
	if len(turn.ToolCalls) == 0 && strings.TrimSpace(turn.Text) == "" {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	return turn, nil
}

type googleConversationContent struct {
	Role  string            `json:"role"`
	Parts []json.RawMessage `json:"parts"`
}
type googleConversationFunctionCall struct {
	ID   string         `json:"id,omitempty"`
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}
type googleConversationPart struct {
	Text         string                          `json:"text,omitempty"`
	Thought      bool                            `json:"thought,omitempty"`
	FunctionCall *googleConversationFunctionCall `json:"functionCall,omitempty"`
}
type googleConversationResponse struct {
	Candidates []struct {
		Content      googleConversationContent `json:"content"`
		FinishReason string                    `json:"finishReason"`
	} `json:"candidates"`
}
type googleConversationDeclaration struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}
type googleConversationWireRequest struct {
	GenerationConfig  *geminiGenerationConfig     `json:"generationConfig"`
	Contents          []googleConversationContent `json:"contents"`
	SystemInstruction *googleConversationContent  `json:"systemInstruction,omitempty"`
	Tools             []googleConversationTool    `json:"tools,omitempty"`
}
type googleConversationTool struct {
	FunctionDeclarations []googleConversationDeclaration `json:"functionDeclarations"`
}

func googleConversationRequest(input ports.ConversationModelInput) (googleConversationWireRequest, error) {
	request := googleConversationWireRequest{GenerationConfig: &geminiGenerationConfig{Temperature: 0}}
	if len(input.Messages) == 0 {
		return request, ports.ErrInvalidProviderInput
	}
	if strings.TrimSpace(input.Instructions) != "" {
		part, _ := json.Marshal(map[string]string{"text": input.Instructions})
		request.SystemInstruction = &googleConversationContent{Role: "user", Parts: []json.RawMessage{part}}
	}
	var declarations []googleConversationDeclaration
	for _, tool := range input.Tools {
		if strings.TrimSpace(tool.Name) == "" || !json.Valid(tool.Parameters) {
			return request, ports.ErrInvalidProviderInput
		}
		declarations = append(declarations, googleConversationDeclaration{Name: tool.Name, Description: tool.Description, Parameters: tool.Parameters})
	}
	if len(declarations) > 0 {
		request.Tools = []googleConversationTool{{FunctionDeclarations: declarations}}
	}
	nativeIDs := map[string]bool{}
	for _, message := range input.Messages {
		if len(message.ProviderState) > googleConversationMaxStateBytes {
			return request, ports.ErrInvalidProviderInput
		}
		if message.Role == ports.ConversationRoleAssistant && len(message.ProviderState) > 0 {
			var original googleConversationContent
			if json.Unmarshal(message.ProviderState, &original) != nil {
				return request, ports.ErrInvalidProviderInput
			}
			for _, raw := range original.Parts {
				var part googleConversationPart
				if json.Unmarshal(raw, &part) != nil {
					return request, ports.ErrInvalidProviderInput
				}
				if part.FunctionCall != nil && part.FunctionCall.ID != "" {
					nativeIDs[part.FunctionCall.ID] = true
				}
			}
		}
		content, err := googleConversationMessage(message, nativeIDs)
		if err != nil {
			return request, err
		}
		// A tool batch is sent as one user content, after the entire model content.
		if message.Role == ports.ConversationRoleTool && len(request.Contents) > 0 && request.Contents[len(request.Contents)-1].Role == "user" {
			previous := &request.Contents[len(request.Contents)-1]
			previous.Parts = append(previous.Parts, content.Parts...)
		} else {
			request.Contents = append(request.Contents, content)
		}
	}
	return request, nil
}

func googleConversationMessage(message ports.ConversationMessage, nativeIDs map[string]bool) (googleConversationContent, error) {
	content := googleConversationContent{Role: "user"}
	if message.Role == ports.ConversationRoleAssistant {
		content.Role = "model"
		if len(message.ProviderState) > 0 {
			if json.Unmarshal(message.ProviderState, &content) != nil || content.Role != "model" || len(content.Parts) == 0 {
				return content, ports.ErrInvalidProviderInput
			}
			return content, nil
		}
	} else if message.Role != ports.ConversationRoleUser && message.Role != ports.ConversationRoleTool {
		return content, ports.ErrInvalidProviderInput
	}
	if message.Text != "" {
		raw, _ := json.Marshal(map[string]string{"text": message.Text})
		content.Parts = append(content.Parts, raw)
	}
	for _, call := range message.ToolCalls {
		raw, err := json.Marshal(googleConversationPart{FunctionCall: &googleConversationFunctionCall{Name: call.Name, Args: call.Arguments}})
		if err != nil {
			return content, ports.ErrInvalidProviderInput
		}
		content.Parts = append(content.Parts, raw)
	}
	for _, result := range message.ToolResults {
		var output any
		if json.Unmarshal([]byte(result.Content), &output) != nil {
			output = result.Content
		}
		functionResponse := map[string]any{"name": result.Name, "response": map[string]any{"output": output}}
		if nativeIDs[result.CallID] {
			functionResponse["id"] = result.CallID
		}
		raw, err := json.Marshal(map[string]any{"functionResponse": functionResponse})
		if err != nil {
			return content, ports.ErrInvalidProviderInput
		}
		content.Parts = append(content.Parts, raw)
	}
	if len(content.Parts) == 0 {
		return content, ports.ErrInvalidProviderInput
	}
	return content, nil
}
