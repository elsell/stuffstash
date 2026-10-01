package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

const compatibleMaxBodyBytes = 1024 * 1024

type compatibleConversation struct {
	endpoint, model, credential string
	client                      *http.Client
}
type compatibleFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
type compatibleCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function compatibleFunction `json:"function"`
}
type compatibleMessage struct {
	Role    string           `json:"role"`
	Content string           `json:"content"`
	Calls   []compatibleCall `json:"tool_calls,omitempty"`
	CallID  string           `json:"tool_call_id,omitempty"`
	Refusal string           `json:"refusal,omitempty"`
}
type compatibleTool struct {
	Type     string               `json:"type"`
	Function compatibleDefinition `json:"function"`
}
type compatibleDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}
type compatibleRequest struct {
	Model    string              `json:"model"`
	Messages []compatibleMessage `json:"messages"`
	Tools    []compatibleTool    `json:"tools,omitempty"`
	Choice   string              `json:"tool_choice,omitempty"`
	Stream   bool                `json:"stream"`
}

func (p compatibleConversation) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	request, err := compatibleWireRequest(p.model, input)
	if err != nil {
		return ports.ConversationModelTurn{}, err
	}
	body, err := json.Marshal(request)
	if err != nil || len(body) > compatibleMaxBodyBytes {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	req.Header.Set("Authorization", "Bearer "+p.credential)
	req.Header.Set("Content-Type", "application/json")
	response, err := p.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ports.ConversationModelTurn{}, ctx.Err()
		}
		return ports.ConversationModelTurn{}, errors.New("compatible provider request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ports.ConversationModelTurn{}, errors.New("compatible provider returned an unsuccessful response")
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, compatibleMaxBodyBytes+1))
	if ctx.Err() != nil {
		return ports.ConversationModelTurn{}, ctx.Err()
	}
	if err != nil || len(raw) > compatibleMaxBodyBytes {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	var result struct {
		Choices []struct {
			Finish  string            `json:"finish_reason"`
			Message compatibleMessage `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(raw, &result) != nil || len(result.Choices) != 1 {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	choice := result.Choices[0]
	message := choice.Message
	if (choice.Finish != "stop" && choice.Finish != "tool_calls") || message.Refusal != "" || (message.Role != "" && message.Role != "assistant") {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	turn := ports.ConversationModelTurn{Text: message.Content}
	names := map[string]bool{}
	for _, tool := range input.Tools {
		names[tool.Name] = true
	}
	ids := map[string]bool{}
	for _, call := range message.Calls {
		var args map[string]any
		if call.Type != "function" || strings.TrimSpace(call.ID) == "" || ids[call.ID] || !names[call.Function.Name] || json.Unmarshal([]byte(call.Function.Arguments), &args) != nil || args == nil {
			return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
		}
		ids[call.ID] = true
		turn.ToolCalls = append(turn.ToolCalls, ports.AgentToolCall{ID: call.ID, Name: call.Function.Name, Arguments: args})
	}
	if len(turn.ToolCalls) == 0 && (request.Choice == "required" || strings.TrimSpace(turn.Text) == "" || choice.Finish == "tool_calls") {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	return turn, nil
}
func compatibleWireRequest(model string, input ports.ConversationModelInput) (compatibleRequest, error) {
	request := compatibleRequest{Model: model}
	if len(input.Messages) == 0 {
		return request, ports.ErrInvalidProviderInput
	}
	if input.Instructions != "" {
		request.Messages = append(request.Messages, compatibleMessage{Role: "system", Content: input.Instructions})
	}
	names := map[string]bool{}
	for _, tool := range input.Tools {
		if strings.TrimSpace(tool.Name) == "" || names[tool.Name] || !json.Valid(tool.Parameters) {
			return request, ports.ErrInvalidProviderInput
		}
		names[tool.Name] = true
		request.Tools = append(request.Tools, compatibleTool{Type: "function", Function: compatibleDefinition{Name: tool.Name, Description: tool.Description, Parameters: tool.Parameters}})
		if tool.ResponseTool {
			request.Choice = "required"
		}
	}
	if len(request.Tools) > 0 && request.Choice == "" {
		request.Choice = "auto"
	}
	pending := map[string]string{}
	for _, message := range input.Messages {
		if message.Role == ports.ConversationRoleTool {
			if len(message.ToolResults) == 0 {
				return request, ports.ErrInvalidProviderInput
			}
			for _, result := range message.ToolResults {
				if name, ok := pending[result.CallID]; !ok || name != result.Name {
					return request, ports.ErrInvalidProviderInput
				}
				delete(pending, result.CallID)
				request.Messages = append(request.Messages, compatibleMessage{Role: "tool", CallID: result.CallID, Content: result.Content})
			}
			continue
		}
		if len(pending) > 0 || (message.Role != ports.ConversationRoleUser && message.Role != ports.ConversationRoleAssistant) {
			return request, ports.ErrInvalidProviderInput
		}
		wire := compatibleMessage{Role: string(message.Role), Content: message.Text}
		if len(message.ToolCalls) > 0 && message.Role != ports.ConversationRoleAssistant {
			return request, ports.ErrInvalidProviderInput
		}
		for _, call := range message.ToolCalls {
			arguments, err := json.Marshal(call.Arguments)
			if err != nil || call.Arguments == nil || strings.TrimSpace(call.ID) == "" || pending[call.ID] != "" {
				return request, ports.ErrInvalidProviderInput
			}
			pending[call.ID] = call.Name
			wire.Calls = append(wire.Calls, compatibleCall{ID: call.ID, Type: "function", Function: compatibleFunction{Name: call.Name, Arguments: string(arguments)}})
		}
		request.Messages = append(request.Messages, wire)
	}
	if len(pending) > 0 {
		return request, ports.ErrInvalidProviderInput
	}
	return request, nil
}
func (p compatibleConversation) ProbeLanguageInference(ctx context.Context) error {
	input := ports.ConversationModelInput{Instructions: "Provider diagnostic. Call ready with status ready.", Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Check tool support."}}, Tools: []ports.ConversationToolDefinition{{Name: "ready", ResponseTool: true, Parameters: json.RawMessage(`{"type":"object","properties":{"status":{"type":"string","enum":["ready"]}},"required":["status"],"additionalProperties":false}`)}}}
	turn, err := p.Converse(ctx, input)
	if err != nil {
		return err
	}
	if len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "ready" || turn.ToolCalls[0].Arguments["status"] != "ready" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}
