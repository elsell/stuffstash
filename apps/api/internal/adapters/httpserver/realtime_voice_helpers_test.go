package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type capturingLanguageModel struct {
	lastToolResult string
}

func (m *capturingLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return typedVoiceInvestigationTurn(input, voiceReadIntent(agentmodel.OperationLocate, "tools"), &m.lastToolResult)
}

func newSeededTestAppWithVoice(t *testing.T, state seededState, stt ports.SpeechToTextProvider, lm ports.LanguageInferenceProvider, tts ports.TextToSpeechProvider) app.App {
	t.Helper()

	application := newSeededTestApp(t, state)
	return application.WithRealtimeVoiceProviders(stt, lm, tts)
}

func seedVoiceAsset(t *testing.T, application app.App, principalID string, tenantID string, inventoryID string, kind string, title string, parentAssetID string) {
	t.Helper()

	_, err := application.CreateAssetWithOperation(context.Background(), app.CreateAssetInput{
		Principal:     identity.Principal{ID: identity.PrincipalID(principalID)},
		Source:        audit.SourceAPI,
		RequestID:     "seed-" + title,
		TenantID:      tenant.ID(tenantID),
		InventoryID:   inventory.InventoryID(inventoryID),
		Kind:          kind,
		Title:         title,
		ParentAssetID: parentAssetID,
	})
	if err != nil {
		t.Fatalf("seed asset %q: %v", title, err)
	}
}

func runRealtimeVoiceQuestion(t *testing.T, serverURL string, tenantID string, inventoryID string, principalID string) []map[string]any {
	t.Helper()

	return runRealtimeVoiceQuestionUntil(t, serverURL, tenantID, inventoryID, principalID, "session.completed")
}

func runRealtimeVoiceQuestionUntil(t *testing.T, serverURL string, tenantID string, inventoryID string, principalID string, terminalType string) []map[string]any {
	t.Helper()

	return runRealtimeVoiceQuestionUntilWithStart(t, serverURL, realtimeVoiceStartMessage(tenantID, inventoryID), principalID, terminalType)
}

func realtimeVoiceStartMessage(tenantID string, inventoryID string) map[string]any {
	return map[string]any{
		"type":        "session.start",
		"seq":         1,
		"tenantId":    tenantID,
		"inventoryId": inventoryID,
		"source":      "mobile_voice",
		"requestedCapabilities": []string{
			"speech_to_text",
			"language_inference",
			"text_to_speech",
		},
		"inputAudio":  map[string]any{"mimeType": "audio/mp4", "sampleRate": 44100, "channels": 1},
		"outputAudio": map[string]any{"mimeTypes": []string{"audio/mpeg"}},
	}
}

func runRealtimeVoiceQuestionUntilWithStart(t *testing.T, serverURL string, startMessage map[string]any, principalID string, terminalType string) []map[string]any {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(serverURL, "http")+"/v1/realtime/voice", &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:" + principalID}},
	})
	if err != nil {
		t.Fatalf("dial realtime voice websocket: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close(websocket.StatusNormalClosure, "") })

	writeRealtimeMessage(t, ctx, connection, startMessage)
	started := readRealtimeMessage(t, ctx, connection)
	sessionID, _ := started["sessionId"].(string)
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":         "audio.chunk",
		"seq":          2,
		"sessionId":    sessionID,
		"chunkId":      "chunk-1",
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio")),
		"isFinalChunk": true,
	})
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.end", "seq": 3, "sessionId": sessionID})
	return readRealtimeMessagesUntil(t, ctx, connection, terminalType)
}

type fakeSpeechToText struct {
	transcript string
	err        error
}

func (f fakeSpeechToText) Transcribe(_ context.Context, input ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	if len(input.AudioChunks) == 0 {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	if f.err != nil {
		return ports.SpeechToTextResult{}, f.err
	}
	return ports.SpeechToTextResult{Transcript: f.transcript}, nil
}

type scriptedSpeechToText struct {
	transcripts []string
}

func (s *scriptedSpeechToText) Transcribe(_ context.Context, input ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	if len(input.AudioChunks) == 0 {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	if len(s.transcripts) == 0 {
		return ports.SpeechToTextResult{}, ports.ErrInvalidProviderInput
	}
	transcript := s.transcripts[0]
	s.transcripts = s.transcripts[1:]
	return ports.SpeechToTextResult{Transcript: transcript}, nil
}

type scriptedLanguageModel struct{}

func (scriptedLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return typedVoiceInvestigationTurn(input, voiceReadIntent(agentmodel.OperationLocate, "tools"), nil)
}

type locationAwareLanguageModel struct {
	lastToolResult string
}

func (m *locationAwareLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return typedVoiceInvestigationTurn(input, voiceReadIntent(agentmodel.OperationLocate, "water bottle"), &m.lastToolResult)
}

type itemListingLanguageModel struct {
	lastToolResult string
}

func (m *itemListingLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	intent := voiceReadIntent(agentmodel.OperationListInventory, "items")
	return typedVoiceInvestigationTurn(input, intent, &m.lastToolResult)
}

type finalResponseLanguageModel struct {
	final ports.StructuredAgentResponse
}

func (m finalResponseLanguageModel) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{Investigation: &agentmodel.InvestigationStep{}}, nil
}

type scriptedFinalLanguageModel struct {
	inputs          []ports.LanguageInferenceInput
	alwaysAmbiguous bool
}

func (m *scriptedFinalLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	m.inputs = append(m.inputs, input)
	if !m.alwaysAmbiguous && len(input.ConversationTurns) > 0 {
		return typedVoiceInvestigationTurn(input, voiceReadIntent(agentmodel.OperationLocate, "Office"), nil)
	}
	return typedAmbiguousItemInvestigationTurn(input)
}

type failingLanguageModel struct {
	err error
}

func (m failingLanguageModel) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{}, m.err
}

func voiceReadIntent(operation agentmodel.Operation, subject string) agentmodel.Intent {
	return agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: operation, SubjectMention: subject}
}

func typedVoiceInvestigationTurn(input ports.LanguageInferenceInput, intent agentmodel.Intent, capture *string) (ports.LanguageInferenceTurn, error) {
	if input.Investigation == nil || input.Investigation.Validate() != nil {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if input.Investigation.Phase == agentmodel.InvestigationPhaseInitial {
		requests := typedVoiceInvestigationRequests(intent)
		step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: requests}
		return ports.LanguageInferenceTurn{Investigation: &step}, nil
	}
	if capture != nil {
		payload, err := json.Marshal(input.Investigation.Observations)
		if err != nil {
			return ports.LanguageInferenceTurn{}, err
		}
		*capture = string(payload)
	}
	resolutions := typedVoiceInvestigationResolutions(intent, input.Investigation.Observations)
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: resolutions}
	return ports.LanguageInferenceTurn{Investigation: &step}, nil
}

func typedVoiceInvestigationRequests(intent agentmodel.Intent) []agentmodel.SearchRequest {
	if intent.Operation == agentmodel.OperationListInventory {
		return []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadListInventory, Mention: intent.SubjectMention, KindHint: "item"}}
	}
	subjectLifecycle := agentmodel.LifecycleScopeActive
	if intent.Operation == agentmodel.OperationRestore {
		subjectLifecycle = agentmodel.LifecycleScopeArchived
	}
	requests := []agentmodel.SearchRequest{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets,
		Mention: intent.SubjectMention, SearchProbes: []string{intent.SubjectMention}, LifecycleScope: subjectLifecycle,
	}}
	for index, segment := range intent.DestinationPath {
		key, _ := agentmodel.NewSemanticReferenceKey("destination." + strconv.Itoa(index))
		requests = append(requests, agentmodel.SearchRequest{ReferenceKey: key, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: segment, SearchProbes: []string{segment}})
	}
	return requests
}

func typedVoiceInvestigationResolutions(intent agentmodel.Intent, observations []agentmodel.CandidateObservation) []agentmodel.Resolution {
	byReference := map[agentmodel.SemanticReferenceKey][]string{}
	for _, observation := range observations {
		byReference[observation.ReferenceKey] = append(byReference[observation.ReferenceKey], observation.CandidateID)
	}
	keys := []agentmodel.SemanticReferenceKey{agentmodel.SemanticReferenceSubject}
	for index := range intent.DestinationPath {
		key, _ := agentmodel.NewSemanticReferenceKey("destination." + strconv.Itoa(index))
		keys = append(keys, key)
	}
	resolutions := make([]agentmodel.Resolution, 0, len(keys))
	for _, key := range keys {
		ids := byReference[key]
		status := agentmodel.ResolutionStrong
		switch {
		case intent.Operation == agentmodel.OperationListInventory && key == agentmodel.SemanticReferenceSubject:
			status = agentmodel.ResolutionCollection
		case len(ids) > 1:
			status = agentmodel.ResolutionAmbiguous
		case len(ids) == 0 && intent.Kind == agentmodel.IntentKindChange && (intent.Operation == agentmodel.OperationCreate || key != agentmodel.SemanticReferenceSubject):
			status = agentmodel.ResolutionMissing
		case len(ids) == 0:
			status = agentmodel.ResolutionAbsent
		}
		resolutions = append(resolutions, agentmodel.Resolution{ReferenceKey: key, Status: status, CandidateIDs: ids, Evidence: "Derived from the authorized test read."})
	}
	return resolutions
}

func typedAmbiguousItemInvestigationTurn(input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	intent := voiceReadIntent(agentmodel.OperationLocate, "item")
	if input.Investigation == nil || input.Investigation.Validate() != nil {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	if input.Investigation.Phase == agentmodel.InvestigationPhaseInitial {
		step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: []agentmodel.SearchRequest{{
			ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadListInventory, Mention: "item", KindHint: "item",
		}}}
		return ports.LanguageInferenceTurn{Investigation: &step}, nil
	}
	ids := make([]string, 0, len(input.Investigation.Observations))
	for _, observation := range input.Investigation.Observations {
		ids = append(ids, observation.CandidateID)
	}
	status := agentmodel.ResolutionAmbiguous
	if len(ids) < 2 {
		status = agentmodel.ResolutionAbsent
		ids = nil
	}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{
		ReferenceKey: agentmodel.SemanticReferenceSubject, Status: status, CandidateIDs: ids, Evidence: "Derived from the authorized test read.",
	}}}
	return ports.LanguageInferenceTurn{Investigation: &step}, nil
}

type fakeTextToSpeech struct {
	chunks [][]byte
}

func (f fakeTextToSpeech) Synthesize(_ context.Context, input ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	if input.Text == "" {
		return ports.TextToSpeechResult{}, ports.ErrInvalidProviderInput
	}
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: f.chunks}, nil
}

func writeRealtimeMessage(t *testing.T, ctx context.Context, connection *websocket.Conn, message map[string]any) {
	t.Helper()

	payload, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal realtime message: %v", err)
	}
	if err := connection.Write(ctx, websocket.MessageText, payload); err != nil {
		t.Fatalf("write realtime message: %v", err)
	}
}

func writeRealtimeAudioTurn(t *testing.T, ctx context.Context, connection *websocket.Conn, sessionID string, seq int, chunkID string) {
	t.Helper()

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":         "audio.chunk",
		"seq":          seq,
		"sessionId":    sessionID,
		"chunkId":      chunkID,
		"audioBase64":  base64.StdEncoding.EncodeToString([]byte("fake-audio-" + chunkID)),
		"isFinalChunk": true,
	})
	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "audio.end",
		"seq":       seq + 1,
		"sessionId": sessionID,
	})
}

func readRealtimeMessage(t *testing.T, ctx context.Context, connection *websocket.Conn) map[string]any {
	t.Helper()

	messageType, payload, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("read realtime message: %v", err)
	}
	if messageType != websocket.MessageText {
		t.Fatalf("expected text message, got %v", messageType)
	}
	var message map[string]any
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatalf("decode realtime message %s: %v", string(payload), err)
	}
	return message
}

func readRealtimeMessagesUntil(t *testing.T, ctx context.Context, connection *websocket.Conn, messageType string) []map[string]any {
	t.Helper()

	var events []map[string]any
	for {
		frameType, payload, err := connection.Read(ctx)
		if err != nil {
			t.Fatalf("read realtime message before %s: %v; events=%+v", messageType, err, events)
		}
		if frameType != websocket.MessageText {
			t.Fatalf("expected text message before %s, got %v; events=%+v", messageType, frameType, events)
		}
		var event map[string]any
		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatalf("decode realtime message %s before %s: %v; events=%+v", string(payload), messageType, err, events)
		}
		events = append(events, event)
		if event["type"] == messageType {
			return events
		}
	}
}

func assertRealtimeEventTypes(t *testing.T, events []map[string]any, expected ...string) {
	t.Helper()

	for _, eventType := range expected {
		if findRealtimeEvent(t, events, eventType) == nil {
			t.Fatalf("expected event type %q in %+v", eventType, events)
		}
	}
}

func assertNoRealtimeEventType(t *testing.T, events []map[string]any, unexpected string) {
	t.Helper()

	for _, event := range events {
		if event["type"] == unexpected {
			t.Fatalf("did not expect event type %q in %+v", unexpected, events)
		}
	}
}

func findRealtimeEvent(t *testing.T, events []map[string]any, eventType string) map[string]any {
	t.Helper()

	for _, event := range events {
		if event["type"] == eventType {
			return event
		}
	}
	return nil
}

func countRealtimeEvents(events []map[string]any, eventType string) int {
	count := 0
	for _, event := range events {
		if event["type"] == eventType {
			count++
		}
	}
	return count
}

func assertSafeRealtimeEvents(t *testing.T, events []map[string]any, forbidden []string) {
	t.Helper()

	payload, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal events: %v", err)
	}
	serialized := string(payload)
	for _, value := range forbidden {
		if strings.Contains(serialized, value) {
			t.Fatalf("realtime events leaked %q: %s", value, serialized)
		}
	}
}
