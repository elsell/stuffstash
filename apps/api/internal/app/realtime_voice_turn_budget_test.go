package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"testing"
	"time"

	appmodel "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type turnBudgetConversationModel struct {
	calls     int
	deadlines []time.Duration
}

func (m *turnBudgetConversationModel) Converse(ctx context.Context, _ ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	if deadline, ok := ctx.Deadline(); ok {
		m.deadlines = append(m.deadlines, time.Until(deadline))
	}
	return ports.ConversationModelTurn{Text: "What would you like to find?"}, nil
}
func TestVoiceWorkflowProcessingBudgetResetsForEachUserTurn(t *testing.T) {
	language := &turnBudgetConversationModel{}
	resolver := &fakeRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{LanguageInferenceProfileID: "lm-profile", SpeechToText: resolvedSpeechToText{transcript: "Hello"}, ConversationModel: language, TextToSpeech: &resolvedTextToSpeech{}}}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{ToolCalls: 2, ModelCalls: 1, ElapsedSeconds: 30, FollowUpTurns: 4}, MaxNameRunes: 100, MaxInstructionRunes: 1000}
	application.conversationWorkflowService = appmodel.NewConversationWorkflowService(appmodel.ConversationWorkflowDependencies{Authorizer: application.authorizer, Repository: store, Profiles: store, IDs: application.ids, Clock: application.clock, Limits: limits})
	seedSessionWorkflow(t, application, store, limits)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	for turn := 0; turn < 3; turn++ {
		err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
		if err != nil {
			t.Fatalf("turn %d consumed a prior utterance's processing budget: %v", turn+1, err)
		}
		if !RealtimeVoiceCanContinue(session) {
			t.Fatal("completed turn disabled permitted followups")
		}
	}
	if language.calls != 3 || len(language.deadlines) != 3 {
		t.Fatalf("model calls/deadlines: %+v", language)
	}
	for _, remaining := range language.deadlines {
		if remaining <= 0 || remaining > 30*time.Second {
			t.Fatalf("turn escaped processing deadline: %s", remaining)
		}
	}
}

type deadlineSessionRepository struct {
	*fakeRealtimeSessionRepository
	cleanupBounded bool
}

func (r *deadlineSessionRepository) UpdateRealtimeSessionOutcome(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, sessionID string, outcome ports.RealtimeSessionOutcome) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, r.cleanupBounded = ctx.Deadline()
	return r.fakeRealtimeSessionRepository.UpdateRealtimeSessionOutcome(ctx, tenantID, inventoryID, sessionID, outcome)
}

type deadlineSpeechToText struct{}

func (deadlineSpeechToText) Transcribe(ctx context.Context, _ ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	<-ctx.Done()
	return ports.SpeechToTextResult{}, ctx.Err()
}
func TestVoiceProcessingDeadlinePersistsTerminalFailure(t *testing.T) {
	language := &turnBudgetConversationModel{}
	resolver := &fakeRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{LanguageInferenceProfileID: "lm-profile", SpeechToText: deadlineSpeechToText{}, ConversationModel: language, TextToSpeech: &resolvedTextToSpeech{}}}
	sessions := &deadlineSessionRepository{fakeRealtimeSessionRepository: newFakeRealtimeSessionRepository()}
	application, store := newRealtimeVoiceResolutionTestAppWithStoreAndSessions(t, resolver, sessions)
	limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{ToolCalls: 2, ModelCalls: 1, ElapsedSeconds: 1, FollowUpTurns: 4}, MaxNameRunes: 100, MaxInstructionRunes: 1000}
	application.conversationWorkflowService = appmodel.NewConversationWorkflowService(appmodel.ConversationWorkflowDependencies{Authorizer: application.authorizer, Repository: store, Profiles: store, IDs: application.ids, Clock: application.clock, Limits: limits})
	seedSessionWorkflow(t, application, store, limits)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	// The outer bound prevents a broken workflow deadline from hanging the suite.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = application.RunRealtimeVoiceQuery(ctx, RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if !errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil || language.calls != 0 {
		t.Fatalf("configured deadline not enforced before inference: %v", err)
	}
	record := sessions.savedRecord(t, session.ID)
	if record.State != ports.RealtimeSessionStateFailed || record.SafeFailureCode != "speech_to_text_failed" || !sessions.cleanupBounded {
		t.Fatalf("failure not persisted with bounded cleanup: %+v", record)
	}
}

type twoReadsConversationModel struct{ calls int }

func (m *twoReadsConversationModel) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	m.calls++
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{
		{ID: "first-search", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "clothes"}},
		{ID: "second-search", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "tools"}},
	}}, nil
}
func TestVoiceWorkflowToolBudgetBoundsModelSelectedReads(t *testing.T) {
	language := &twoReadsConversationModel{}
	speech := &resolvedTextToSpeech{}
	resolver := &fakeRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{LanguageInferenceProfileID: "lm-profile", SpeechToText: resolvedSpeechToText{transcript: "Find clothes and tools"}, ConversationModel: language, TextToSpeech: speech}}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{ToolCalls: 1, ModelCalls: 4, ElapsedSeconds: 30, FollowUpTurns: 4}, MaxNameRunes: 100, MaxInstructionRunes: 1000}
	application.conversationWorkflowService = appmodel.NewConversationWorkflowService(appmodel.ConversationWorkflowDependencies{Authorizer: application.authorizer, Repository: store, Profiles: store, IDs: application.ids, Clock: application.clock, Limits: limits})
	seedSessionWorkflow(t, application, store, limits)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if !errors.Is(err, appmodel.ErrConversationBudgetExhausted) || language.calls != 1 || speech.lastText != "" {
		t.Fatalf("configured tool cap ignored: calls=%d err=%v", language.calls, err)
	}
}
