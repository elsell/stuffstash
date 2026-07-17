package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type realtimeVoiceInvestigationReadState struct {
	seenQueries  map[agentmodel.SemanticReferenceKey]map[string]struct{}
	visible      map[agentmodel.SemanticReferenceKey]map[string]agentmodel.CandidateObservation
	toolResults  []ports.AgentToolResult
	toolCallIDs  []string
	readEvidence []agentmodel.ReadEvidence
}

type realtimeVoiceInvestigationReadResult struct {
	Observations []agentmodel.CandidateObservation
	ToolResults  []ports.AgentToolResult
	ToolCallIDs  []string
	ReadEvidence []agentmodel.ReadEvidence
}

func newRealtimeVoiceInvestigationReadState(previous []agentmodel.SearchRequest, observations []agentmodel.CandidateObservation, readEvidence []agentmodel.ReadEvidence) (*realtimeVoiceInvestigationReadState, error) {
	state := &realtimeVoiceInvestigationReadState{
		seenQueries: map[agentmodel.SemanticReferenceKey]map[string]struct{}{},
		visible:     map[agentmodel.SemanticReferenceKey]map[string]agentmodel.CandidateObservation{},
	}
	for _, request := range previous {
		if request.Validate() != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		for _, probe := range request.SearchProbes {
			state.recordQuery(request.ReferenceKey, probe, request.LifecycleScope)
		}
	}
	for _, observation := range observations {
		if observation.Validate() != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		state.mergeObservation(observation)
	}
	for _, evidence := range readEvidence {
		if evidence.Validate() != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		state.readEvidence = append(state.readEvidence, evidence)
	}
	return state, nil
}

func (a App) executeRealtimeVoiceInvestigationReads(ctx context.Context, session RealtimeVoiceSession, evidenceRound int, requests []agentmodel.SearchRequest, state *realtimeVoiceInvestigationReadState, emit RealtimeVoiceEventSink) (realtimeVoiceInvestigationReadResult, error) {
	if state == nil || evidenceRound < 1 || evidenceRound > agentmodel.MaxEvidenceRounds || len(requests) == 0 || len(requests) > agentmodel.MaxSearchRequestsPerStep {
		return realtimeVoiceInvestigationReadResult{}, ports.ErrInvalidProviderInput
	}
	newObservations := map[string]agentmodel.CandidateObservation{}
	startResults := len(state.toolResults)
	startIDs := len(state.toolCallIDs)
	startEvidence := len(state.readEvidence)
	if err := emitRealtimeVoiceProgress(session, realtimeVoiceProgressExploring, "Checking your inventory.", emit); err != nil {
		return realtimeVoiceInvestigationReadResult{}, err
	}
	for _, request := range requests {
		if request.Validate() != nil {
			return realtimeVoiceInvestigationReadResult{}, ports.ErrInvalidProviderInput
		}
		calls, err := a.realtimeVoiceInvestigationCalls(request, state)
		if err != nil {
			return realtimeVoiceInvestigationReadResult{}, err
		}
		for _, callSpec := range calls {
			call := callSpec.call
			call.ID = a.newRealtimeVoiceID()
			if strings.TrimSpace(call.ID) == "" {
				return realtimeVoiceInvestigationReadResult{}, ports.ErrInvalidProviderInput
			}
			label := realtimeVoiceToolLabel(call.Name)
			state.toolCallIDs = append(state.toolCallIDs, call.ID)
			if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallStarted, SessionID: session.ID, ToolCallID: call.ID, ToolLabel: label, Status: "searching"}); err != nil {
				return realtimeVoiceInvestigationReadResult{}, err
			}
			visibleIDs := state.visibleIDs(request.ReferenceKey)
			result, _, err := a.executeRealtimeVoiceTool(ctx, session, "", state.toolResults, call, visibleIDs)
			if err != nil {
				_ = emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallFailed, SessionID: session.ID, ToolCallID: call.ID, ToolLabel: label, Code: "invalid_tool_request", Message: "I could not check that safely."})
				return realtimeVoiceInvestigationReadResult{}, err
			}
			state.toolResults = append(state.toolResults, result)
			observations, err := realtimeVoiceInvestigationObservationsFromToolResult(evidenceRound, request.ReferenceKey, callSpec.probe, result)
			if err != nil {
				return realtimeVoiceInvestigationReadResult{}, err
			}
			state.readEvidence = append(state.readEvidence, agentmodel.ReadEvidence{
				EvidenceRound: evidenceRound, ReferenceKey: request.ReferenceKey, ReadKind: request.ReadKind,
				Probe: strings.TrimSpace(callSpec.probe), VisibleAssetID: strings.TrimSpace(request.VisibleAssetID), CandidateCount: len(observations), LifecycleScope: request.LifecycleScope.Effective(),
			})
			for _, observation := range observations {
				key := observation.ReferenceKey.String() + "\x00" + observation.CandidateID
				if existing, exists := newObservations[key]; exists {
					observation = mergeRealtimeVoiceInvestigationObservation(existing, observation)
				}
				newObservations[key] = observation
				state.mergeObservation(observation)
			}
			if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallCompleted, SessionID: session.ID, ToolCallID: call.ID, ToolLabel: label, Status: realtimeVoiceToolCompletionStatus(result)}); err != nil {
				return realtimeVoiceInvestigationReadResult{}, err
			}
		}
	}
	observations := make([]agentmodel.CandidateObservation, 0, len(newObservations))
	for _, observation := range newObservations {
		observations = append(observations, observation)
	}
	return realtimeVoiceInvestigationReadResult{
		Observations: observations,
		ToolResults:  append([]ports.AgentToolResult{}, state.toolResults[startResults:]...),
		ToolCallIDs:  append([]string{}, state.toolCallIDs[startIDs:]...),
		ReadEvidence: append([]agentmodel.ReadEvidence{}, state.readEvidence[startEvidence:]...),
	}, nil
}

type realtimeVoiceInvestigationCall struct {
	call  ports.AgentToolCall
	probe string
}

func (a App) realtimeVoiceInvestigationCalls(request agentmodel.SearchRequest, state *realtimeVoiceInvestigationReadState) ([]realtimeVoiceInvestigationCall, error) {
	switch request.ReadKind {
	case agentmodel.InvestigationReadSearchAssets:
		calls := make([]realtimeVoiceInvestigationCall, 0, len(request.SearchProbes))
		for _, probe := range request.SearchProbes {
			if state.querySeen(request.ReferenceKey, probe, request.LifecycleScope) {
				return nil, ports.ErrInvalidProviderInput
			}
			state.recordQuery(request.ReferenceKey, probe, request.LifecycleScope)
			calls = append(calls, realtimeVoiceInvestigationCall{probe: strings.TrimSpace(probe), call: ports.AgentToolCall{Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": strings.TrimSpace(probe), "lifecycleState": string(request.LifecycleScope.Effective()), "limit": realtimeVoiceToolMaxResults}}})
		}
		return calls, nil
	case agentmodel.InvestigationReadListInventory:
		arguments := map[string]any{"limit": realtimeVoiceToolMaxResults, "lifecycleState": string(request.LifecycleScope.Effective())}
		if request.KindHint != "" {
			arguments["kind"] = request.KindHint
		}
		return []realtimeVoiceInvestigationCall{{call: ports.AgentToolCall{Name: RealtimeVoiceToolListAuthorizedAssets, Arguments: arguments}}}, nil
	case agentmodel.InvestigationReadListContents:
		candidate, ok := state.visible[request.ReferenceKey][request.VisibleAssetID]
		if !ok || (candidate.Kind != "location" && candidate.Kind != "container") {
			return nil, ports.ErrInvalidProviderInput
		}
		arguments := map[string]any{"limit": realtimeVoiceToolMaxResults, "lifecycleState": string(request.LifecycleScope.Effective()), "parentAssetId": request.VisibleAssetID}
		return []realtimeVoiceInvestigationCall{{call: ports.AgentToolCall{Name: RealtimeVoiceToolListAuthorizedAssets, Arguments: arguments}}}, nil
	case agentmodel.InvestigationReadAssetDetail:
		if !state.assetVisibleForReference(request.ReferenceKey, request.VisibleAssetID) {
			return nil, ports.ErrInvalidProviderInput
		}
		return []realtimeVoiceInvestigationCall{{call: ports.AgentToolCall{Name: RealtimeVoiceToolGetAssetDetail, Arguments: map[string]any{"assetId": request.VisibleAssetID}}}}, nil
	case agentmodel.InvestigationReadAssetHistory:
		if !state.assetVisibleForReference(request.ReferenceKey, request.VisibleAssetID) {
			return nil, ports.ErrInvalidProviderInput
		}
		return []realtimeVoiceInvestigationCall{{call: ports.AgentToolCall{Name: RealtimeVoiceToolListAssetAuditHistory, Arguments: map[string]any{"assetId": request.VisibleAssetID, "limit": realtimeVoiceToolMaxResults}}}}, nil
	case agentmodel.InvestigationReadCheckoutHistory:
		if !state.assetVisibleForReference(request.ReferenceKey, request.VisibleAssetID) {
			return nil, ports.ErrInvalidProviderInput
		}
		return []realtimeVoiceInvestigationCall{{call: ports.AgentToolCall{Name: RealtimeVoiceToolListAssetCheckoutHistory, Arguments: map[string]any{"assetId": request.VisibleAssetID, "limit": realtimeVoiceToolMaxResults}}}}, nil
	default:
		return nil, ports.ErrInvalidProviderInput
	}
}

func realtimeVoiceInvestigationObservationsFromToolResult(round int, reference agentmodel.SemanticReferenceKey, probe string, result ports.AgentToolResult) ([]agentmodel.CandidateObservation, error) {
	switch result.Name {
	case RealtimeVoiceToolSearchAuthorizedAssets, RealtimeVoiceToolListAuthorizedAssets, RealtimeVoiceToolGetAssetDetail:
		var output realtimeVoiceAssetToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		observations := make([]agentmodel.CandidateObservation, 0, len(output.Items))
		for _, item := range output.Items {
			observations = append(observations, realtimeVoiceInvestigationObservationFromItem(round, reference, probe, item, nil))
		}
		return observations, nil
	case RealtimeVoiceToolListAssetAuditHistory:
		var output realtimeVoiceAssetAuditHistoryToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		facts := make([]string, 0, len(output.Entries))
		for _, entry := range output.Entries {
			facts = append(facts, entry.Summary)
		}
		return []agentmodel.CandidateObservation{realtimeVoiceInvestigationObservationFromItem(round, reference, probe, output.Asset, facts)}, nil
	case RealtimeVoiceToolListAssetCheckoutHistory:
		var output realtimeVoiceAssetCheckoutHistoryToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			return nil, ports.ErrInvalidProviderInput
		}
		facts := make([]string, 0, len(output.Entries))
		for _, entry := range output.Entries {
			fact := "Checked out at " + entry.CheckedOutAt
			if entry.ReturnedAt != "" {
				fact += " and returned at " + entry.ReturnedAt
			}
			facts = append(facts, fact)
		}
		return []agentmodel.CandidateObservation{realtimeVoiceInvestigationObservationFromItem(round, reference, probe, output.Asset, facts)}, nil
	default:
		return nil, ports.ErrInvalidProviderInput
	}
}

func realtimeVoiceInvestigationObservationFromItem(round int, reference agentmodel.SemanticReferenceKey, probe string, item realtimeVoiceAssetToolItem, facts []string) agentmodel.CandidateObservation {
	checkoutState := "available"
	if item.CurrentCheckout != nil || (item.CheckoutState != nil && item.CheckoutState.CheckedOut) {
		checkoutState = "checked_out"
	}
	matched := []string{}
	if strings.TrimSpace(probe) != "" {
		matched = append(matched, strings.TrimSpace(probe))
	}
	return agentmodel.CandidateObservation{
		EvidenceRound: round, ReferenceKey: reference, CandidateID: item.AssetID, Title: item.Title, Kind: item.Kind,
		Description: item.Description, ParentAssetID: item.ParentAssetID, LifecycleState: item.LifecycleState,
		CheckoutState: checkoutState, ContainmentPath: append([]string{}, item.ContainmentPath...), MatchedProbes: matched, Facts: facts,
	}
}

func mergeRealtimeVoiceInvestigationObservation(left, right agentmodel.CandidateObservation) agentmodel.CandidateObservation {
	merged := right
	merged.MatchedProbes = appendUniqueRealtimeVoiceInvestigation(append([]string{}, left.MatchedProbes...), right.MatchedProbes...)
	merged.Facts = appendUniqueRealtimeVoiceInvestigation(append([]string{}, left.Facts...), right.Facts...)
	if merged.Description == "" {
		merged.Description = left.Description
	}
	return merged
}

func (state *realtimeVoiceInvestigationReadState) mergeObservation(observation agentmodel.CandidateObservation) {
	if state.visible[observation.ReferenceKey] == nil {
		state.visible[observation.ReferenceKey] = map[string]agentmodel.CandidateObservation{}
	}
	if existing, ok := state.visible[observation.ReferenceKey][observation.CandidateID]; ok {
		observation = mergeRealtimeVoiceInvestigationObservation(existing, observation)
	}
	state.visible[observation.ReferenceKey][observation.CandidateID] = observation
}

func (state *realtimeVoiceInvestigationReadState) visibleIDs(reference agentmodel.SemanticReferenceKey) map[string]struct{} {
	ids := map[string]struct{}{}
	for id := range state.visible[reference] {
		ids[id] = struct{}{}
	}
	return ids
}

func (state *realtimeVoiceInvestigationReadState) assetVisibleForReference(reference agentmodel.SemanticReferenceKey, id string) bool {
	_, exists := state.visible[reference][strings.TrimSpace(id)]
	return exists
}

func (state *realtimeVoiceInvestigationReadState) querySeen(reference agentmodel.SemanticReferenceKey, query string, lifecycle agentmodel.LifecycleScope) bool {
	_, exists := state.seenQueries[reference][realtimeVoiceInvestigationQueryKey(query, lifecycle)]
	return exists
}

func (state *realtimeVoiceInvestigationReadState) recordQuery(reference agentmodel.SemanticReferenceKey, query string, lifecycle agentmodel.LifecycleScope) {
	if state.seenQueries[reference] == nil {
		state.seenQueries[reference] = map[string]struct{}{}
	}
	state.seenQueries[reference][realtimeVoiceInvestigationQueryKey(query, lifecycle)] = struct{}{}
}

func realtimeVoiceInvestigationQueryKey(query string, lifecycle agentmodel.LifecycleScope) string {
	return string(lifecycle.Effective()) + "\x00" + normalizeRealtimeVoiceInvestigationTitle(query)
}

func appendUniqueRealtimeVoiceInvestigation(values []string, additions ...string) []string {
	for _, addition := range additions {
		addition = strings.TrimSpace(addition)
		if addition == "" {
			continue
		}
		duplicate := false
		for _, value := range values {
			if strings.EqualFold(strings.TrimSpace(value), addition) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			values = append(values, addition)
		}
	}
	return values
}
