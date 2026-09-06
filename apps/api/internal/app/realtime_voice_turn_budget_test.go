package app

import (
	"context"
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
