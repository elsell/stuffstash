package app

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type realtimeConversationTools struct {
	proposal    *RealtimeVoiceActionPlanProposal
	application App
	session     RealtimeVoiceSession
	emit        RealtimeVoiceEventSink
	visible     map[string]struct{}
	items       map[string]realtimeVoiceAssetToolItem
	callIDs     []string
	results     []ports.AgentToolResult
}

func (e *realtimeConversationTools) ExecuteConversationTool(ctx context.Context, call ports.AgentToolCall) (ports.ConversationToolOutcome, error) {
	if err := e.application.ensureRealtimeVoiceAccess(ctx, e.session.Principal, e.session.TenantID, e.session.InventoryID); err != nil {
		return ports.ConversationToolOutcome{}, err
	}
	label := realtimeVoiceToolLabel(call.Name)
	if call.Name == realtimeConversationProposeTool {
		label = "Prepare changes"
	} else if call.Name == realtimeConversationPresentTool {
		label = "Prepare answer"
	}
	if err := e.emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallStarted, SessionID: e.session.ID, ToolCallID: call.ID, ToolLabel: label}); err != nil {
		return ports.ConversationToolOutcome{}, err
	}
	var outcome ports.ConversationToolOutcome
	var result ports.AgentToolResult
	var err error
	if call.Name == realtimeConversationProposeTool {
		outcome, err = e.propose(ctx, call)
		result = outcome.Result
	} else if call.Name == realtimeConversationPresentTool {
		outcome, err = e.present(call)
		result = outcome.Result
	} else {
		result, err = e.application.executeRealtimeVoiceTool(ctx, e.session, call, e.visible)
	}
	if err != nil {
		if ctx.Err() != nil {
			return ports.ConversationToolOutcome{}, ctx.Err()
		}
		if errors.Is(err, ports.ErrForbidden) || errors.Is(err, ports.ErrUnauthenticated) || errors.Is(err, context.Canceled) {
			return ports.ConversationToolOutcome{}, err
		}
		content := `{"error":"This operation is temporarily unavailable. You may retry within the remaining budget or explain the limitation using existing evidence."}`
		if errors.Is(err, ports.ErrInvalidProviderInput) || errors.Is(err, apperrors.ErrInvalidInput) {
			content = `{"error":"Invalid tool arguments or unavailable tool. Check the tool definition and correct the request."}`
		}
		result = ports.AgentToolResult{CallID: call.ID, Name: call.Name, Call: call, Content: content}
	} else {
		var output realtimeVoiceAssetToolOutput
		if json.Unmarshal([]byte(result.Content), &output) == nil {
			for _, item := range output.Items {
				if item.AssetID == "" {
					continue
				}
				e.visible[item.AssetID] = struct{}{}
				e.items[item.AssetID] = item
			}
		}
	}
	result.CallID, result.Name, result.Call = call.ID, call.Name, call
	e.callIDs = append(e.callIDs, call.ID)
	e.results = append(e.results, result)
	eventType, status := RealtimeVoiceEventToolCallCompleted, realtimeVoiceToolCompletionStatus(result)
	if err != nil {
		eventType, status = RealtimeVoiceEventToolCallFailed, "failed"
	}
	if err := e.emit(RealtimeVoiceEvent{Type: eventType, SessionID: e.session.ID, ToolCallID: call.ID, ToolLabel: label, Status: status}); err != nil {
		return ports.ConversationToolOutcome{}, err
	}
	outcome.Result = result
	return outcome, nil
}

func realtimeConversationReadTools() []ports.ConversationToolDefinition {
	return []ports.ConversationToolDefinition{
		realtimeConversationVocabularyTool(),
		{Name: RealtimeVoiceToolSearchAuthorizedAssets, Description: "Search authorized inventory names, descriptions and tags. Every whitespace-separated query term must occur somewhere in the same item; terms may match different fields or tags. Results already include parentTitle and containmentPath for recorded locations; you can answer a location question directly from these results without fetching detail. An empty lexical search does not cover differently named items or related categories. Use distinct category synonyms when useful; relevance must be assessed from the results.", Parameters: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"},"lifecycleState":{"type":"string","enum":["active","archived","all"]},"limit":{"type":"integer","minimum":1,"maximum":20,"description":"Maximum returned candidates; defaults to 20. A low limit can hide competing matches. hasMore means additional matches exist, so a truncated result does not establish uniqueness."}},"required":["query"],"additionalProperties":false}`)},
		{Name: RealtimeVoiceToolGetAssetDetail, Description: "Read additional facts about an asset returned by a prior authorized read, using its exact assetId. Use this when required facts are missing; search already includes recorded location and containment, so repeating a detail read solely for location is unnecessary.", Parameters: json.RawMessage(`{"type":"object","properties":{"assetId":{"type":"string"}},"required":["assetId"],"additionalProperties":false}`)},
		{Name: RealtimeVoiceToolListAuthorizedAssets, Description: "Browse item records in this inventory or the contents of a known container. Useful for discovering belongings when their names or tags are unknown. Inspect the returned titles, descriptions and tags to decide relevance; this tool does not filter by conceptual category. kind restricts results to item, container or location; omit it to include all kinds. hasMore means results are incomplete.", Parameters: json.RawMessage(`{"type":"object","properties":{"kind":{"type":"string","enum":["item","container","location"]},"parentAssetId":{"type":"string"},"lifecycleState":{"type":"string","enum":["active","archived","all"]},"limit":{"type":"integer","minimum":1,"maximum":20}},"additionalProperties":false}`)},
		{Name: RealtimeVoiceToolListAssetAuditHistory, Description: "Read recorded changes to an observed asset, newest first. Use its exact authorized assetId. Entries are history, not proof of current state; read detail when current facts are needed.", Parameters: json.RawMessage(`{"type":"object","properties":{"assetId":{"type":"string"},"limit":{"type":"integer","minimum":1,"maximum":20}},"required":["assetId"],"additionalProperties":false}`)},
		{Name: RealtimeVoiceToolListCheckedOutAssets, Description: "List currently checked-out items in this inventory with their current checkout records. hasMore means coverage is incomplete.", Parameters: json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","minimum":1,"maximum":20}},"additionalProperties":false}`)},
		{Name: RealtimeVoiceToolListAssetCheckoutHistory, Description: "Read checkout and return history for an observed asset using its exact authorized assetId. Results are newest first; hasMore means older entries may exist.", Parameters: json.RawMessage(`{"type":"object","properties":{"assetId":{"type":"string"},"limit":{"type":"integer","minimum":1,"maximum":20}},"required":["assetId"],"additionalProperties":false}`)},
	}
}
