package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceCheckoutHistoryObservationIncludesBoundedCheckoutDetails(t *testing.T) {
	t.Parallel()
	payload, err := json.Marshal(realtimeVoiceAssetCheckoutHistoryToolOutput{
		Asset:   realtimeVoiceAssetToolItem{AssetID: "flashlight-1", Title: "Loaner flashlight", Kind: "item"},
		Entries: []realtimeVoiceAssetCheckoutHistoryEntry{{CheckedOutAt: "2026-06-29T13:00:00Z", CheckoutDetails: "Loaned to Sam"}},
	})
	if err != nil {
		t.Fatalf("marshal checkout history: %v", err)
	}
	observations, err := realtimeVoiceInvestigationObservationsFromToolResult(2, agentmodel.SemanticReferenceSubject, "", ports.AgentToolResult{Name: RealtimeVoiceToolListAssetCheckoutHistory, Content: string(payload)})
	if err != nil || len(observations) != 1 {
		t.Fatalf("map checkout history: %+v, %v", observations, err)
	}
	if len(observations[0].Facts) != 1 || !strings.Contains(observations[0].Facts[0], "Sam") {
		t.Fatalf("expected authorized checkout details in bounded fact, got %+v", observations[0].Facts)
	}
}

func TestRealtimeVoiceInvestigationReadsMergeSearchProbeEvidenceByReference(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	closet := realtimeVoiceInvestigationAsset("closet-1", "Hall closet", asset.KindContainer, "")
	winterClothes := realtimeVoiceInvestigationAsset("winter-clothes-1", "Sarah Winter Clothes and Shoes", asset.KindItem, closet.ID.String())
	seedRealtimeVoiceLoopAsset(t, store, closet, "audit-closet")
	seedRealtimeVoiceLoopAsset(t, store, winterClothes, "audit-winter-clothes")

	state, err := newRealtimeVoiceInvestigationReadState(nil, nil, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	var events []RealtimeVoiceEvent
	result, err := application.executeRealtimeVoiceInvestigationReads(
		context.Background(),
		realtimeVoiceInvestigationSession(),
		1,
		[]agentmodel.SearchRequest{{
			ReferenceKey: agentmodel.SemanticReferenceSubject,
			ReadKind:     agentmodel.InvestigationReadSearchAssets,
			Mention:      "Sarah's winter coat",
			SearchProbes: []string{"Sarah", "winter clothes"},
		}},
		state,
		func(event RealtimeVoiceEvent) error {
			events = append(events, event)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("execute investigation reads: %v", err)
	}
	if len(result.Observations) != 1 {
		t.Fatalf("expected one deduplicated observation, got %+v", result.Observations)
	}
	observation := result.Observations[0]
	if observation.ReferenceKey != agentmodel.SemanticReferenceSubject || observation.CandidateID != winterClothes.ID.String() || observation.ParentAssetID != closet.ID.String() {
		t.Fatalf("unexpected grounded observation: %+v", observation)
	}
	if len(observation.MatchedProbes) != 2 || observation.MatchedProbes[0] != "Sarah" || observation.MatchedProbes[1] != "winter clothes" {
		t.Fatalf("expected both distinct probes to be retained, got %+v", observation.MatchedProbes)
	}
	if len(result.ToolResults) != 2 || len(result.ToolCallIDs) != 2 {
		t.Fatalf("expected one project-owned tool trace per probe, got results=%d ids=%d", len(result.ToolResults), len(result.ToolCallIDs))
	}
	if len(result.ReadEvidence) != 2 || result.ReadEvidence[0].ReferenceKey != agentmodel.SemanticReferenceSubject || result.ReadEvidence[0].CandidateCount != 1 || result.ReadEvidence[1].CandidateCount != 1 {
		t.Fatalf("expected safe-read evidence for both completed probes, got %+v", result.ReadEvidence)
	}
	if !realtimeVoiceInvestigationHasEvent(events, RealtimeVoiceEventAgentProgress) ||
		!realtimeVoiceInvestigationHasEvent(events, RealtimeVoiceEventToolCallStarted) ||
		!realtimeVoiceInvestigationHasEvent(events, RealtimeVoiceEventToolCallCompleted) {
		t.Fatalf("expected safe progress and tool events, got %+v", events)
	}
}

func TestRealtimeVoiceInvestigationReadsRetainZeroMatchEvidence(t *testing.T) {
	t.Parallel()

	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	state, err := newRealtimeVoiceInvestigationReadState(nil, nil, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	result, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), realtimeVoiceInvestigationSession(), 1, []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject,
		ReadKind:     agentmodel.InvestigationReadSearchAssets,
		Mention:      "moon boots",
		SearchProbes: []string{"moon boots"},
	}}, state, func(RealtimeVoiceEvent) error { return nil })
	if err != nil {
		t.Fatalf("execute zero-match search: %v", err)
	}
	if len(result.Observations) != 0 {
		t.Fatalf("expected no candidate observations, got %+v", result.Observations)
	}
	if len(result.ReadEvidence) != 1 || result.ReadEvidence[0].CandidateCount != 0 || result.ReadEvidence[0].Probe != "moon boots" {
		t.Fatalf("expected explicit zero-match read evidence, got %+v", result.ReadEvidence)
	}
}

func TestRealtimeVoiceInvestigationReadsRejectRepeatedProbeOnlyWithinReference(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	item := realtimeVoiceInvestigationAsset("coat-1", "Sarah coat", asset.KindItem, "")
	seedRealtimeVoiceLoopAsset(t, store, item, "audit-coat")
	state, err := newRealtimeVoiceInvestigationReadState(nil, nil, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	session := realtimeVoiceInvestigationSession()
	if _, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), session, 1, []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject,
		ReadKind:     agentmodel.InvestigationReadSearchAssets,
		Mention:      "Sarah coat",
		SearchProbes: []string{"Sarah"},
	}}, state, func(RealtimeVoiceEvent) error { return nil }); err != nil {
		t.Fatalf("execute first probe: %v", err)
	}
	if _, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), session, 2, []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject,
		ReadKind:     agentmodel.InvestigationReadSearchAssets,
		Mention:      "Sarah coat",
		SearchProbes: []string{"  SARAH  "},
	}}, state, func(RealtimeVoiceEvent) error { return nil }); !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected repeated normalized subject probe rejection, got %v", err)
	}

	destination, ok := agentmodel.NewSemanticReferenceKey("destination.0")
	if !ok {
		t.Fatal("expected valid destination reference")
	}
	if _, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), session, 2, []agentmodel.SearchRequest{{
		ReferenceKey: destination,
		ReadKind:     agentmodel.InvestigationReadSearchAssets,
		Mention:      "Sarah room",
		SearchProbes: []string{"sarah"},
	}}, state, func(RealtimeVoiceEvent) error { return nil }); err != nil {
		t.Fatalf("same normalized probe must remain valid for a different semantic reference: %v", err)
	}
}

func TestRealtimeVoiceInvestigationTypedReadsRequireSameReferenceVisibility(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	item := realtimeVoiceInvestigationAsset("drill-1", "Cordless drill", asset.KindItem, "")
	seedRealtimeVoiceLoopAsset(t, store, item, "audit-drill")
	destination, ok := agentmodel.NewSemanticReferenceKey("destination.0")
	if !ok {
		t.Fatal("expected valid destination reference")
	}
	prior := []agentmodel.CandidateObservation{{
		EvidenceRound: 1,
		ReferenceKey:  agentmodel.SemanticReferenceSubject,
		CandidateID:   item.ID.String(),
		Title:         item.Title.String(),
		Kind:          item.Kind.String(),
	}}
	state, err := newRealtimeVoiceInvestigationReadState(nil, prior, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	request := agentmodel.SearchRequest{
		ReferenceKey:   destination,
		ReadKind:       agentmodel.InvestigationReadAssetDetail,
		VisibleAssetID: item.ID.String(),
	}
	if _, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), realtimeVoiceInvestigationSession(), 2, []agentmodel.SearchRequest{request}, state, func(RealtimeVoiceEvent) error { return nil }); !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected cross-reference visible ID rejection, got %v", err)
	}

	request.ReferenceKey = agentmodel.SemanticReferenceSubject
	result, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), realtimeVoiceInvestigationSession(), 2, []agentmodel.SearchRequest{request}, state, func(RealtimeVoiceEvent) error { return nil })
	if err != nil {
		t.Fatalf("execute same-reference detail: %v", err)
	}
	if len(result.Observations) != 1 || result.Observations[0].CandidateID != item.ID.String() {
		t.Fatalf("expected grounded detail observation, got %+v", result.Observations)
	}
}

func TestRealtimeVoiceInvestigationReadsDropsRepeatedProbeWhenStepStillMakesProgress(t *testing.T) {
	t.Parallel()
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("coat-1", "Sarah winter coat", asset.KindItem, ""), "audit-progressive-probe")
	state, err := newRealtimeVoiceInvestigationReadState(nil, nil, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	session := realtimeVoiceInvestigationSession()
	if _, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), session, 1, []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets,
		Mention: "Sarah winter coat", SearchProbes: []string{"Sarah"},
	}}, state, func(RealtimeVoiceEvent) error { return nil }); err != nil {
		t.Fatalf("execute first probe: %v", err)
	}
	result, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), session, 2, []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets,
		Mention: "Sarah winter coat", SearchProbes: []string{" SARAH ", "winter coat"},
	}}, state, func(RealtimeVoiceEvent) error { return nil })
	if err != nil {
		t.Fatalf("expected novel probe to remain valid after deduplication: %v", err)
	}
	if len(result.ReadEvidence) != 1 || result.ReadEvidence[0].Probe != "winter coat" {
		t.Fatalf("expected only the novel probe to execute, got %+v", result.ReadEvidence)
	}
}

func TestRealtimeVoiceInvestigationReadStateClearsDestinationCapabilitiesForRepair(t *testing.T) {
	t.Parallel()
	state, err := newRealtimeVoiceInvestigationReadState(
		[]agentmodel.SearchRequest{
			{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "drill", SearchProbes: []string{"drill"}},
			{ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "rack", SearchProbes: []string{"rack"}},
		},
		[]agentmodel.CandidateObservation{
			{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill-1", Title: "Drill", Kind: "item"},
			{EvidenceRound: 1, ReferenceKey: "destination.0", CandidateID: "rack-1", Title: "Rack", Kind: "container"},
		},
		[]agentmodel.ReadEvidence{
			{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "drill", CandidateCount: 1},
			{EvidenceRound: 1, ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "rack", CandidateCount: 1},
		},
	)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	state.resetDestinationScope()
	if state.querySeen("destination.0", "rack", agentmodel.LifecycleScopeActive) || state.assetVisibleForReference("destination.0", "rack-1") {
		t.Fatal("expected stale destination query and visible ID capability to be cleared")
	}
	if !state.querySeen(agentmodel.SemanticReferenceSubject, "drill", agentmodel.LifecycleScopeActive) || !state.assetVisibleForReference(agentmodel.SemanticReferenceSubject, "drill-1") {
		t.Fatal("expected subject evidence and visibility to remain available")
	}
	if len(state.readEvidence) != 1 || state.readEvidence[0].ReferenceKey != agentmodel.SemanticReferenceSubject {
		t.Fatalf("expected only subject read evidence after repair reset, got %+v", state.readEvidence)
	}
}

func TestRealtimeVoiceInvestigationReadsMapInventoryContentsAndTypedHistory(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	toolbox := realtimeVoiceInvestigationAsset("toolbox-1", "Toolbox", asset.KindContainer, "")
	drill := realtimeVoiceInvestigationAsset("drill-1", "Cordless drill", asset.KindItem, toolbox.ID.String())
	seedRealtimeVoiceLoopAsset(t, store, toolbox, "audit-toolbox")
	seedRealtimeVoiceLoopAsset(t, store, drill, "audit-drill")

	contentsReference, ok := agentmodel.NewSemanticReferenceKey("destination.0")
	if !ok {
		t.Fatal("expected valid contents reference")
	}
	prior := []agentmodel.CandidateObservation{
		{EvidenceRound: 1, ReferenceKey: contentsReference, CandidateID: toolbox.ID.String(), Title: toolbox.Title.String(), Kind: toolbox.Kind.String()},
		{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: drill.ID.String(), Title: drill.Title.String(), Kind: drill.Kind.String(), ParentAssetID: toolbox.ID.String()},
	}
	state, err := newRealtimeVoiceInvestigationReadState(nil, prior, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	requests := []agentmodel.SearchRequest{
		{ReferenceKey: contentsReference, ReadKind: agentmodel.InvestigationReadListContents, VisibleAssetID: toolbox.ID.String()},
		{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadAssetDetail, VisibleAssetID: drill.ID.String()},
		{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadAssetHistory, VisibleAssetID: drill.ID.String()},
		{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadCheckoutHistory, VisibleAssetID: drill.ID.String()},
	}
	result, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), realtimeVoiceInvestigationSession(), 2, requests, state, func(RealtimeVoiceEvent) error { return nil })
	if err != nil {
		t.Fatalf("execute typed reads: %v", err)
	}
	if len(result.ToolResults) != len(requests) {
		t.Fatalf("expected one tool result per typed read, got %+v", result.ToolResults)
	}
	if result.ToolResults[0].Name != RealtimeVoiceToolListAuthorizedAssets || result.ToolResults[1].Name != RealtimeVoiceToolGetAssetDetail ||
		result.ToolResults[2].Name != RealtimeVoiceToolListAssetAuditHistory || result.ToolResults[3].Name != RealtimeVoiceToolListAssetCheckoutHistory {
		t.Fatalf("unexpected typed read mapping: %+v", result.ToolResults)
	}
	contents := realtimeVoiceInvestigationObservation(result.Observations, contentsReference, drill.ID.String())
	if contents == nil || contents.ParentAssetID != toolbox.ID.String() {
		t.Fatalf("expected list-contents child with parent provenance, got %+v", result.Observations)
	}
	history := realtimeVoiceInvestigationObservation(result.Observations, agentmodel.SemanticReferenceSubject, drill.ID.String())
	if history == nil || len(history.Facts) == 0 {
		t.Fatalf("expected deduplicated subject observation to retain typed history facts, got %+v", result.Observations)
	}
}

func TestRealtimeVoiceInvestigationListContentsScopesDuplicateTitlesByVisibleParentID(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	selectedToolbox := realtimeVoiceInvestigationAsset("toolbox-selected", "Toolbox", asset.KindContainer, "")
	otherToolbox := realtimeVoiceInvestigationAsset("toolbox-other", "Toolbox", asset.KindContainer, "")
	selectedDrill := realtimeVoiceInvestigationAsset("drill-selected", "Cordless drill", asset.KindItem, selectedToolbox.ID.String())
	otherDrill := realtimeVoiceInvestigationAsset("drill-other", "Impact drill", asset.KindItem, otherToolbox.ID.String())
	seedRealtimeVoiceLoopAsset(t, store, selectedToolbox, "audit-toolbox-selected")
	seedRealtimeVoiceLoopAsset(t, store, otherToolbox, "audit-toolbox-other")
	seedRealtimeVoiceLoopAsset(t, store, selectedDrill, "audit-drill-selected")
	seedRealtimeVoiceLoopAsset(t, store, otherDrill, "audit-drill-other")

	contentsReference, ok := agentmodel.NewSemanticReferenceKey("destination.0")
	if !ok {
		t.Fatal("expected valid contents reference")
	}
	state, err := newRealtimeVoiceInvestigationReadState(nil, []agentmodel.CandidateObservation{{
		EvidenceRound: 1,
		ReferenceKey:  contentsReference,
		CandidateID:   selectedToolbox.ID.String(),
		Title:         selectedToolbox.Title.String(),
		Kind:          selectedToolbox.Kind.String(),
	}}, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}

	result, err := application.executeRealtimeVoiceInvestigationReads(
		context.Background(),
		realtimeVoiceInvestigationSession(),
		2,
		[]agentmodel.SearchRequest{{
			ReferenceKey:   contentsReference,
			ReadKind:       agentmodel.InvestigationReadListContents,
			VisibleAssetID: selectedToolbox.ID.String(),
		}},
		state,
		func(RealtimeVoiceEvent) error { return nil },
	)
	if err != nil {
		t.Fatalf("execute list contents: %v", err)
	}
	if len(result.Observations) != 1 || result.Observations[0].CandidateID != selectedDrill.ID.String() {
		t.Fatalf("expected only the selected toolbox child, got %+v", result.Observations)
	}
	if strings.Contains(result.ToolResults[0].Content, otherDrill.ID.String()) {
		t.Fatalf("duplicate-titled parent leaked another container's child: %s", result.ToolResults[0].Content)
	}
}

func TestRealtimeVoiceInvestigationReadsMapBroadInventoryList(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	item := realtimeVoiceInvestigationAsset("labels-1", "Freezer labels", asset.KindItem, "")
	seedRealtimeVoiceLoopAsset(t, store, item, "audit-labels")
	state, err := newRealtimeVoiceInvestigationReadState(nil, nil, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	result, err := application.executeRealtimeVoiceInvestigationReads(context.Background(), realtimeVoiceInvestigationSession(), 1, []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject,
		ReadKind:     agentmodel.InvestigationReadListInventory,
		KindHint:     "item",
	}}, state, func(RealtimeVoiceEvent) error { return nil })
	if err != nil {
		t.Fatalf("execute list inventory: %v", err)
	}
	if len(result.ToolResults) != 1 || result.ToolResults[0].Name != RealtimeVoiceToolListAuthorizedAssets || len(result.Observations) != 1 || result.Observations[0].CandidateID != item.ID.String() {
		t.Fatalf("unexpected inventory list result: %+v", result)
	}
}

func realtimeVoiceInvestigationSession() RealtimeVoiceSession {
	input := defaultRealtimeVoiceSessionInput()
	return RealtimeVoiceSession{
		ID:          "investigation-session",
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Principal:   input.Principal,
		Source:      input.Source,
	}
}

func realtimeVoiceInvestigationAsset(id string, title string, kind asset.Kind, parentID string) asset.Asset {
	item := assetItem(id, "tenant-home", "inventory-home", kind, parentID)
	itemTitle, ok := asset.NewTitle(title)
	if !ok {
		panic("invalid investigation test asset title")
	}
	item.Title = itemTitle
	return item
}

func realtimeVoiceInvestigationHasEvent(events []RealtimeVoiceEvent, eventType string) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func realtimeVoiceInvestigationObservation(observations []agentmodel.CandidateObservation, reference agentmodel.SemanticReferenceKey, candidateID string) *agentmodel.CandidateObservation {
	for index := range observations {
		if observations[index].ReferenceKey == reference && observations[index].CandidateID == candidateID {
			return &observations[index]
		}
	}
	return nil
}
