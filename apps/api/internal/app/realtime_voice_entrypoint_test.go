package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceProductionEntrypointSupportsReadOperationMatrix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		operation agentmodel.Operation
		requests  func(asset.ID, asset.ID) []agentmodel.SearchRequest
		status    agentmodel.ResolutionStatus
		candidate func(asset.ID, asset.ID) string
	}{
		{operation: agentmodel.OperationLocate, requests: realtimeVoiceSubjectSearchRequests, status: agentmodel.ResolutionStrong, candidate: realtimeVoiceSubjectCandidate},
		{operation: agentmodel.OperationExists, requests: realtimeVoiceSubjectSearchRequests, status: agentmodel.ResolutionStrong, candidate: realtimeVoiceSubjectCandidate},
		{operation: agentmodel.OperationListInventory, requests: func(_, _ asset.ID) []agentmodel.SearchRequest {
			return []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadListInventory}}
		}, status: agentmodel.ResolutionCollection, candidate: realtimeVoiceSubjectCandidate},
		{operation: agentmodel.OperationListContents, requests: func(_, containerID asset.ID) []agentmodel.SearchRequest {
			return []agentmodel.SearchRequest{
				{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Storage fixture", SearchProbes: []string{"Storage fixture"}},
				{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadListContents, VisibleAssetID: containerID.String()},
			}
		}, status: agentmodel.ResolutionCollection, candidate: realtimeVoiceSubjectCandidate},
		{operation: agentmodel.OperationDetail, requests: func(subjectID, _ asset.ID) []agentmodel.SearchRequest {
			return append(realtimeVoiceSubjectSearchRequests(subjectID, ""), agentmodel.SearchRequest{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadAssetDetail, VisibleAssetID: subjectID.String()})
		}, status: agentmodel.ResolutionStrong, candidate: realtimeVoiceSubjectCandidate},
		{operation: agentmodel.OperationCheckoutStatus, requests: func(subjectID, _ asset.ID) []agentmodel.SearchRequest {
			return append(realtimeVoiceSubjectSearchRequests(subjectID, ""), agentmodel.SearchRequest{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadCheckoutHistory, VisibleAssetID: subjectID.String()})
		}, status: agentmodel.ResolutionStrong, candidate: realtimeVoiceSubjectCandidate},
		{operation: agentmodel.OperationAssetHistory, requests: func(subjectID, _ asset.ID) []agentmodel.SearchRequest {
			return append(realtimeVoiceSubjectSearchRequests(subjectID, ""), agentmodel.SearchRequest{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadAssetHistory, VisibleAssetID: subjectID.String()})
		}, status: agentmodel.ResolutionStrong, candidate: realtimeVoiceSubjectCandidate},
		{operation: agentmodel.OperationCheckoutHistory, requests: func(subjectID, _ asset.ID) []agentmodel.SearchRequest {
			return append(realtimeVoiceSubjectSearchRequests(subjectID, ""), agentmodel.SearchRequest{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadCheckoutHistory, VisibleAssetID: subjectID.String()})
		}, status: agentmodel.ResolutionStrong, candidate: realtimeVoiceSubjectCandidate},
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(string(testCase.operation), func(t *testing.T) {
			t.Parallel()
			resolver := successfulRealtimeVoiceResolver()
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: fmt.Sprintf("generated read request: %s", testCase.operation)}
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			container := realtimeVoiceInvestigationAsset("matrix-container", "Storage fixture", asset.KindContainer, "")
			subject := realtimeVoiceInvestigationAsset("matrix-subject", "Subject record", asset.KindItem, container.ID.String())
			seedRealtimeVoiceLoopAsset(t, store, container, "audit-matrix-container")
			seedRealtimeVoiceLoopAsset(t, store, subject, "audit-matrix-subject")

			intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: testCase.operation, SubjectMention: "Subject record"}
			if testCase.operation == agentmodel.OperationListInventory {
				intent.SubjectMention = ""
			}
			if testCase.operation == agentmodel.OperationListContents {
				intent.SubjectMention = "Storage fixture"
			}
			initial := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: testCase.requests(subject.ID, container.ID)}
			final := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
				ReferenceKey: agentmodel.SemanticReferenceSubject,
				Status:       testCase.status,
				CandidateIDs: []string{testCase.candidate(subject.ID, container.ID)},
			}}}
			language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &final}}}
			resolver.providers.LanguageInference = language

			events := runRealtimeVoiceProductionEntrypoint(t, application)
			if response := realtimeVoiceInvestigationCompletedResponse(events); response == nil || response.Kind != ports.StructuredAgentResponseKindAnswer {
				t.Fatalf("expected a grounded answer from production entrypoint, got %+v", events)
			}
			if !realtimeVoiceInvestigationHasEvent(events, RealtimeVoiceEventSessionCompleted) {
				t.Fatalf("expected completed session, got %+v", events)
			}
			if got := resolver.providers.TextToSpeech.(*resolvedTextToSpeech).lastText; got == "" {
				t.Fatal("expected grounded response to reach text-to-speech")
			}
			if len(language.seenInvestigations) != 2 || language.seenInvestigations[0].Phase != agentmodel.InvestigationPhaseInitial || language.seenInvestigations[1].Phase != agentmodel.InvestigationPhaseEvidenceAssessment {
				t.Fatalf("expected bounded initial and evidence-assessment turns, got %+v", language.seenInvestigations)
			}
		})
	}
}

func TestRealtimeVoiceProductionEntrypointSupportsChangeOperationMatrix(t *testing.T) {
	t.Parallel()

	for _, operation := range []agentmodel.Operation{
		agentmodel.OperationCreate,
		agentmodel.OperationMove,
		agentmodel.OperationArchive,
		agentmodel.OperationRestore,
		agentmodel.OperationCheckout,
		agentmodel.OperationReturn,
	} {
		operation := operation
		t.Run(string(operation), func(t *testing.T) {
			t.Parallel()
			resolver := successfulRealtimeVoiceResolver()
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: fmt.Sprintf("generated change request: %s", operation)}
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			subject := realtimeVoiceInvestigationAsset("change-subject", "Change subject", asset.KindItem, "")
			destination := realtimeVoiceInvestigationAsset("change-destination", "Destination fixture", asset.KindLocation, "")
			if operation == agentmodel.OperationRestore {
				subject.LifecycleState = asset.LifecycleStateArchived
			}
			if operation != agentmodel.OperationCreate {
				seedRealtimeVoiceLoopAsset(t, store, subject, "audit-change-subject")
			}
			if operation == agentmodel.OperationMove {
				seedRealtimeVoiceLoopAsset(t, store, destination, "audit-change-destination")
			}
			if operation == agentmodel.OperationReturn {
				seedRealtimeVoiceOpenCheckout(t, store, subject.ID)
			}

			intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: operation, SubjectMention: "Change subject"}
			requests := []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Change subject", SearchProbes: []string{"Change subject"}}}
			resolutions := []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{subject.ID.String()}}}
			switch operation {
			case agentmodel.OperationCreate:
				intent.NewAssetKind = asset.KindItem.String()
				resolutions[0] = agentmodel.Resolution{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionMissing}
			case agentmodel.OperationMove:
				intent.DestinationPath = []string{"Destination fixture"}
				intent.DestinationKinds = []agentmodel.DestinationKind{agentmodel.DestinationKindLocation}
				requests = append(requests, agentmodel.SearchRequest{ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Destination fixture", SearchProbes: []string{"Destination fixture"}})
				resolutions = append(resolutions, agentmodel.Resolution{ReferenceKey: "destination.0", Status: agentmodel.ResolutionStrong, CandidateIDs: []string{destination.ID.String()}})
			case agentmodel.OperationRestore:
				requests = []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadListInventory, Mention: "Change subject", KindHint: asset.KindItem.String(), LifecycleScope: agentmodel.LifecycleScopeArchived}}
			}

			initial := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: requests}
			final := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: resolutions}
			language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &final}}}
			resolver.providers.LanguageInference = language

			events := runRealtimeVoiceProductionEntrypoint(t, application)
			proposal := realtimeVoiceInvestigationProposedPlan(events)
			if proposal == nil || len(proposal.Commands) == 0 {
				t.Fatalf("expected reviewable %s proposal from production entrypoint, got %+v", operation, events)
			}
			if realtimeVoiceInvestigationHasEvent(events, RealtimeVoiceEventAssistantResponseCompleted) || realtimeVoiceInvestigationHasEvent(events, RealtimeVoiceEventSessionCompleted) {
				t.Fatalf("write must pause at review without completing or speaking, got %+v", events)
			}
			if got := resolver.providers.TextToSpeech.(*resolvedTextToSpeech).lastText; got != "" {
				t.Fatalf("write proposal must not reach text-to-speech before approval, got %q", got)
			}
		})
	}
}

func TestRealtimeVoiceProductionEntrypointHandlesAmbiguousAndAbsentEvidence(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		status     agentmodel.ResolutionStatus
		seedTitles []string
		expected   ports.StructuredAgentResponseKind
	}{
		{name: "ambiguous", status: agentmodel.ResolutionAmbiguous, seedTitles: []string{"Matching record alpha", "Matching record beta"}, expected: ports.StructuredAgentResponseKindClarification},
		{name: "absent", status: agentmodel.ResolutionAbsent, expected: ports.StructuredAgentResponseKindAnswer},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			resolver := successfulRealtimeVoiceResolver()
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "generated evidence-state request"}
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			ids := []string{}
			for index, title := range testCase.seedTitles {
				item := realtimeVoiceInvestigationAsset(fmt.Sprintf("evidence-%d", index), title, asset.KindItem, "")
				seedRealtimeVoiceLoopAsset(t, store, item, fmt.Sprintf("audit-evidence-%d", index))
				ids = append(ids, item.ID.String())
			}
			intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "Matching record"}
			initial := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: []agentmodel.SearchRequest{{
				ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Matching record", SearchProbes: []string{"Matching record"},
			}}}
			final := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
				ReferenceKey: agentmodel.SemanticReferenceSubject, Status: testCase.status, CandidateIDs: ids,
			}}}
			resolver.providers.LanguageInference = &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &final}}}

			events := runRealtimeVoiceProductionEntrypoint(t, application)
			if response := realtimeVoiceInvestigationCompletedResponse(events); response == nil || response.Kind != testCase.expected {
				t.Fatalf("expected %s terminal response, got %+v", testCase.expected, events)
			}
		})
	}
}

func TestRealtimeVoiceProductionEntrypointRejectsMalformedStructuredTurn(t *testing.T) {
	t.Parallel()

	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "generated malformed-output request"}
	resolver.providers.LanguageInference = &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{}}}
	application := newRealtimeVoiceResolutionTestApp(t, resolver)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if err == nil {
		t.Fatal("expected model output outside the investigation schema to be rejected")
	}
	if got := resolver.providers.TextToSpeech.(*resolvedTextToSpeech).lastText; got != "" {
		t.Fatalf("malformed model output must not reach speech, got %q", got)
	}
}

func TestRealtimeVoiceProductionEntrypointBoundsInvestigationReads(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "Timed subject"}
	initial := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Timed subject", SearchProbes: []string{"Timed subject"},
	}}}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "generated bounded-read request"}
	resolver.providers.LanguageInference = language
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	search := &blockingAssetSearchRepository{ready: make(chan struct{})}
	application.search = search
	application.realtimeVoiceToolCallTimeout = time.Millisecond
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err == nil {
		t.Fatal("expected bounded investigation read timeout to terminate the request")
	}
	if !search.cancelled {
		t.Fatal("expected investigation read context to be cancelled at the tool boundary")
	}
	if !realtimeVoiceToolTimeoutEvent(events, "invalid_tool_request") {
		t.Fatalf("expected safe failed-read event, got %+v", events)
	}
}

func TestRealtimeVoiceInvestigationReadsDoNotMaskParentDeadline(t *testing.T) {
	t.Parallel()
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	application.search = &blockingAssetSearchRepository{ready: make(chan struct{})}
	application.realtimeVoiceToolCallTimeout = time.Minute
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err = application.executeRealtimeVoiceTool(ctx, session, ports.AgentToolCall{
		ID: "bounded-search", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Timed subject"},
	}, map[string]struct{}{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected parent deadline to remain terminal, got %T %[1]v", err)
	}
}

func realtimeVoiceSubjectSearchRequests(_, _ asset.ID) []agentmodel.SearchRequest {
	return []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject,
		ReadKind:     agentmodel.InvestigationReadSearchAssets,
		Mention:      "Subject record",
		SearchProbes: []string{"Subject record"},
	}}
}

func realtimeVoiceSubjectCandidate(subjectID, _ asset.ID) string { return subjectID.String() }

func runRealtimeVoiceProductionEntrypoint(t *testing.T, application App) []RealtimeVoiceEvent {
	t.Helper()
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run production realtime voice entrypoint: %v", err)
	}
	return events
}

func seedRealtimeVoiceOpenCheckout(t *testing.T, store interface {
	CheckOutAsset(context.Context, asset.Checkout, audit.Record, *ports.UndoableOperation) error
}, assetID asset.ID) {
	t.Helper()
	now := time.Date(2026, 6, 29, 13, 8, 30, 0, time.UTC)
	if err := store.CheckOutAsset(context.Background(), asset.Checkout{
		ID: asset.CheckoutID("matrix-checkout"), TenantID: asset.TenantID("tenant-home"), InventoryID: asset.InventoryID("inventory-home"),
		AssetID: assetID, State: asset.CheckoutStateOpen, CheckedOutAt: now, CheckedOutByPrincipal: "user-1", CreatedAt: now, UpdatedAt: now,
	}, audit.Record{
		ID: audit.ID("audit-matrix-checkout"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"),
		Action: audit.ActionAssetCheckedOut, TargetType: audit.TargetAsset, TargetID: assetID.String(), OccurredAt: now,
	}, nil); err != nil {
		t.Fatalf("seed open checkout: %v", err)
	}
}

type blockingAssetSearchRepository struct {
	ready     chan struct{}
	cancelled bool
}

func (r *blockingAssetSearchRepository) SearchAssets(ctx context.Context, _ tenant.ID, _ []inventory.InventoryID, _ ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	close(r.ready)
	<-ctx.Done()
	r.cancelled = true
	return nil, ctx.Err()
}

func realtimeVoiceToolTimeoutEvent(events []RealtimeVoiceEvent, code string) bool {
	for _, event := range events {
		if event.Type == RealtimeVoiceEventToolCallFailed && event.Code == code {
			return true
		}
	}
	return false
}
