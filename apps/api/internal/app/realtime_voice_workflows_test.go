package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	appmodel "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeSessionPinsWorkflowAndEnforcesPerTurnModelLimit(t *testing.T) {
	language := &resolvedConversationModel{inventoryConversationModel: inventoryConversationModel{query: "tools"}}
	resolver := &fakeRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{LanguageInferenceProfileID: "lm-profile", SpeechToText: resolvedSpeechToText{transcript: "Where are my tools?"}, ConversationModel: language, TextToSpeech: &resolvedTextToSpeech{}}}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{ToolCalls: 2, ModelCalls: 4, ElapsedSeconds: 60, FollowUpTurns: 4}, MaxNameRunes: 100, MaxInstructionRunes: 1000}
	application.conversationWorkflowService = appmodel.NewConversationWorkflowService(appmodel.ConversationWorkflowDependencies{Authorizer: application.authorizer, Repository: store, Profiles: store, IDs: application.ids, Clock: application.clock, Limits: limits})
	revision := seedSessionWorkflow(t, application, store, limits)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	if session.WorkflowRevisionID != string(revision.Snapshot().ID) {
		t.Fatalf("selected workflow not pinned: %+v", session)
	}
	// A new selected revision must not change the session's per-turn call allowance.
	input := SaveConversationWorkflowInput{Principal: session.Principal, TenantID: session.TenantID, Source: audit.SourceAPI, WorkflowID: revision.Snapshot().WorkflowID, ExpectedRevision: 1, Definition: revision.Snapshot().Definition.Settings()}
	input.Definition.Budget.ModelCalls = 1
	updated, err := application.SaveConversationWorkflowRevision(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	activateSessionWorkflow(t, application, store, updated, ports.WorkflowSelectionReference{WorkflowID: revision.Snapshot().WorkflowID, RevisionID: revision.Snapshot().ID})
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if err != nil {
		t.Fatalf("existing session was redirected: %v", err)
	}
	if language.calls != 2 {
		t.Fatalf("expected model search and answer: %d", language.calls)
	}
	newer, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	if newer.WorkflowRevisionID != string(updated.Snapshot().ID) {
		t.Fatal("new session did not select new revision")
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: newer, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if err == nil || language.calls != 3 {
		t.Fatalf("new session exceeded one model call: calls=%d err=%v", language.calls, err)
	}
}

func seedSessionWorkflow(t *testing.T, application App, store *memory.Store, limits agentmodel.WorkflowLimits) agentmodel.WorkflowRevision {
	t.Helper()
	session := defaultRealtimeVoiceSessionInput()
	revision, err := application.SaveConversationWorkflowRevision(context.Background(), SaveConversationWorkflowInput{Principal: session.Principal, TenantID: session.TenantID, Source: audit.SourceAPI, Definition: agentmodel.WorkflowDefinitionInput{Name: "Voice", Budget: limits.Budget}})
	if err != nil {
		t.Fatal(err)
	}
	activateSessionWorkflow(t, application, store, revision, ports.WorkflowSelectionReference{})
	return revision
}
func activateSessionWorkflow(t *testing.T, application App, store *memory.Store, revision agentmodel.WorkflowRevision, expected ports.WorkflowSelectionReference) {
	t.Helper()
	snapshot := revision.Snapshot()
	record, ok := audit.NewRecord(audit.ID(application.ids.NewID()), audit.TenantID(snapshot.TenantID), "", audit.PrincipalID(snapshot.AuthorID), audit.ActionConversationWorkflowActivated, audit.SourceAPI, audit.TargetType("conversation_workflow"), string(snapshot.WorkflowID), application.clock.Now(), "", nil)
	if !ok {
		t.Fatal("invalid activation audit")
	}
	if err := store.ActivateWorkflowRevision(context.Background(), defaultRealtimeVoiceSessionInput().TenantID, snapshot.WorkflowID, snapshot.ID, expected, application.clock.Now(), record); err != nil {
		t.Fatal(err)
	}
}
