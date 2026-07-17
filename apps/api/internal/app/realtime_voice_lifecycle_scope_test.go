package app

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func TestRealtimeVoiceInvestigationAllowsSameProbeAcrossDifferentLifecycleScopesOnly(t *testing.T) {
	t.Parallel()
	state, err := newRealtimeVoiceInvestigationReadState(nil, nil, nil)
	if err != nil {
		t.Fatalf("new read state: %v", err)
	}
	application := App{}
	request := agentmodel.SearchRequest{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "drill", SearchProbes: []string{"drill"}, LifecycleScope: agentmodel.LifecycleScopeActive}
	if _, err := application.realtimeVoiceInvestigationCalls(request, state); err != nil {
		t.Fatalf("active read: %v", err)
	}
	request.LifecycleScope = agentmodel.LifecycleScopeArchived
	if _, err := application.realtimeVoiceInvestigationCalls(request, state); err != nil {
		t.Fatalf("archived read after active miss must be distinct: %v", err)
	}
	if _, err := application.realtimeVoiceInvestigationCalls(request, state); err == nil {
		t.Fatal("expected exact same probe and lifecycle scope to remain a duplicate")
	}
}

func TestCanonicalRealtimeVoiceInvestigationRejectsCandidateOutsideDiscoveryLifecycleScope(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "old drill"}
	step := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"drill-1"}}}}
	observations := []agentmodel.CandidateObservation{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, CandidateID: "drill-1", Title: "Old drill", Kind: "item", LifecycleState: "archived"}}
	evidence := []agentmodel.ReadEvidence{{EvidenceRound: 1, ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Probe: "old drill", CandidateCount: 1, LifecycleScope: agentmodel.LifecycleScopeActive}}
	if _, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, evidence); err == nil {
		t.Fatal("expected active-only evidence to reject an archived candidate")
	}
	evidence[0].LifecycleScope = agentmodel.LifecycleScopeAll
	if _, err := canonicalRealtimeVoiceInvestigationStep(intent, step, observations, evidence); err != nil {
		t.Fatalf("all lifecycle evidence should ground archived candidate: %v", err)
	}
}
