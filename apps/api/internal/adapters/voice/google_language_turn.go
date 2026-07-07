package voice

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func parseLanguageTurn(raw string, tools []ports.AgentToolDescriptor, allowActionPlan bool) (ports.LanguageInferenceTurn, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	var decoded languageTurnJSON
	decoder := json.NewDecoder(bytes.NewReader([]byte(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	if decoder.More() {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if decoded.Final != nil && (len(decoded.ToolCalls) > 0 || decoded.ActionPlan != nil) {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if decoded.ActionPlan != nil && len(decoded.ToolCalls) > 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if decoded.ActionPlan != nil && !allowActionPlan {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if decoded.Final != nil {
		kind := ports.StructuredAgentResponseKind(decoded.Final.Kind)
		if !isAllowedStructuredResponseKind(kind) ||
			!boundedNonEmpty(decoded.Final.SpokenResponse, 500) ||
			!boundedOptional(decoded.Final.DisplayResponse, 1000) {
			return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
		}
		return ports.LanguageInferenceTurn{Final: &ports.StructuredAgentResponse{
			Kind:            kind,
			SpokenResponse:  decoded.Final.SpokenResponse,
			DisplayResponse: decoded.Final.DisplayResponse,
		}}, nil
	}
	if decoded.ActionPlan != nil {
		if !validLanguageActionPlanEnvelope(decoded.ActionPlan) {
			return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
		}
		return ports.LanguageInferenceTurn{ToolCalls: []ports.AgentToolCall{{
			ID:        "gemini-action-plan",
			Name:      "propose_action_plan",
			Arguments: decoded.ActionPlan,
		}}}, nil
	}
	allowedTools := allowedToolNames(tools)
	toolCalls := make([]ports.AgentToolCall, 0, len(decoded.ToolCalls))
	for _, call := range decoded.ToolCalls {
		if !boundedNonEmpty(call.ID, 100) || !allowedTools[call.Name] || call.Arguments == nil {
			return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
		}
		toolCalls = append(toolCalls, ports.AgentToolCall{
			ID:        call.ID,
			Name:      call.Name,
			Arguments: call.Arguments,
		})
	}
	if len(toolCalls) == 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.LanguageInferenceTurn{ToolCalls: toolCalls}, nil
}

func validLanguageActionPlanEnvelope(plan map[string]any) bool {
	if !boundedNonEmpty(stringValue(plan["intentSummary"]), 500) ||
		!boundedNonEmpty(stringValue(plan["modelInterpretationSummary"]), 1000) ||
		!boundedNonEmpty(stringValue(plan["confirmationSummary"]), 500) {
		return false
	}
	commands, ok := plan["commands"].([]any)
	if !ok || len(commands) == 0 {
		return false
	}
	for _, raw := range commands {
		command, ok := raw.(map[string]any)
		if !ok ||
			!boundedNonEmpty(stringValue(command["id"]), 100) ||
			!allowedLanguageActionPlanCommandKind(stringValue(command["kind"])) ||
			!boundedNonEmpty(stringValue(command["summary"]), 500) {
			return false
		}
		if _, ok := command["arguments"].(map[string]any); !ok {
			return false
		}
	}
	return true
}

func allowedLanguageActionPlanCommandKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "create_asset", "create_location", "move_asset", "archive_asset", "restore_asset", "checkout_asset", "return_asset":
		return true
	default:
		return false
	}
}

type languageTurnJSON struct {
	ToolCalls  []languageToolCallJSON `json:"toolCalls,omitempty"`
	Final      *languageFinalJSON     `json:"final,omitempty"`
	ActionPlan map[string]any         `json:"actionPlan,omitempty"`
}

type languageToolCallJSON struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type languageFinalJSON struct {
	Kind            string `json:"kind"`
	SpokenResponse  string `json:"spokenResponse"`
	DisplayResponse string `json:"displayResponse"`
}
