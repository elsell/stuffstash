package app

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeVoiceInvestigationVersion = "voice-investigation-v1"

func (a App) runRealtimeVoiceInvestigationLoop(ctx context.Context, session RealtimeVoiceSession, transcript string, conversationTurns []ports.AgentConversationTurn, continueAfterClarification bool, emit RealtimeVoiceEventSink) error {
	vocabulary, vocabularyCatalog, err := a.loadRealtimeVoiceVocabulary(ctx, session.TenantID, session.InventoryID)
	if err != nil {
		return err
	}
	initialInput := agentmodel.InvestigationInput{
		Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: realtimeVoiceInvestigationVersion,
		SchemaVersion: realtimeVoiceInvestigationVersion, Transcript: transcript, MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
		Vocabulary: vocabulary,
	}
	step, err := a.nextRealtimeVoiceInvestigation(ctx, session, transcript, conversationTurns, initialInput, nil, emit)
	if err != nil {
		return err
	}
	vocabularyDefinitions, err := vocabularyCatalog.resolve(step.VocabularyRequests)
	if err != nil {
		return a.recoverRealtimeVoiceResponse(ctx, session, nil, nil, emit)
	}
	vocabularyRequests := append([]agentmodel.VoiceVocabularyRequest{}, step.VocabularyRequests...)
	if step.Decision == agentmodel.InvestigationDecisionFinish {
		if step.Intent.Kind != agentmodel.IntentKindUnsupported {
			return a.recoverRealtimeVoiceResponse(ctx, session, nil, nil, emit)
		}
		canonical, err := canonicalRealtimeVoiceInvestigationStep(step.Intent, step, nil, nil)
		if err != nil {
			return a.recoverRealtimeVoiceResponse(ctx, session, nil, nil, emit)
		}
		response, err := realtimeVoiceInvestigationResponse(canonical.Intent, canonical.Resolutions, nil)
		if err != nil {
			return a.recoverRealtimeVoiceResponse(ctx, session, nil, nil, emit)
		}
		return a.completeRealtimeVoiceResponse(ctx, session, response, nil, nil, emit, continueAfterClarification)
	}
	if step.Decision != agentmodel.InvestigationDecisionSearch {
		return a.recoverRealtimeVoiceResponse(ctx, session, nil, nil, emit)
	}
	canonicalIntent := step.Intent
	if canonicalIntent.Validate() != nil || canonicalIntent.Kind == agentmodel.IntentKindUnsupported {
		return a.recoverRealtimeVoiceResponse(ctx, session, nil, nil, emit)
	}
	readState, err := newRealtimeVoiceInvestigationReadState(nil, nil, nil)
	if err != nil {
		return err
	}
	requests := []agentmodel.SearchRequest{}
	observations := []agentmodel.CandidateObservation{}
	readEvidence := []agentmodel.ReadEvidence{}
	for evidenceRound := 1; evidenceRound <= agentmodel.MaxEvidenceRounds; evidenceRound++ {
		readResult, err := a.executeRealtimeVoiceInvestigationReads(ctx, session, evidenceRound, step.SearchRequests, readState, emit)
		if err != nil {
			return err
		}
		requests = append(requests, step.SearchRequests...)
		observations = mergeRealtimeVoiceInvestigationObservations(observations, readResult.Observations)
		readEvidence = append(readEvidence, readResult.ReadEvidence...)
		intentCopy := canonicalIntent
		assessmentInput := agentmodel.InvestigationInput{
			Phase: agentmodel.InvestigationPhaseEvidenceAssessment, PromptVersion: realtimeVoiceInvestigationVersion,
			SchemaVersion: realtimeVoiceInvestigationVersion, Transcript: transcript, EvidenceRound: evidenceRound,
			MaxEvidenceRounds: agentmodel.MaxEvidenceRounds, CanonicalIntent: &intentCopy,
			PreviousRequests: append([]agentmodel.SearchRequest{}, requests...), Observations: append([]agentmodel.CandidateObservation{}, observations...),
			ReadEvidence: append([]agentmodel.ReadEvidence{}, readEvidence...),
			Vocabulary:   vocabulary, VocabularyRequests: append([]agentmodel.VoiceVocabularyRequest{}, vocabularyRequests...),
			VocabularyDefinitions: append([]agentmodel.VoiceVocabularyDefinition{}, vocabularyDefinitions...),
		}
		step, err = a.nextRealtimeVoiceInvestigation(ctx, session, transcript, conversationTurns, assessmentInput, readState.toolResults, emit)
		if err != nil {
			return err
		}
		if !sameRealtimeVoiceInvestigationIntent(canonicalIntent, step.Intent) {
			return a.recoverRealtimeVoiceResponse(ctx, session, readState.toolCallIDs, readState.toolResults, emit)
		}
		if step.Decision == agentmodel.InvestigationDecisionSearchAgain {
			if evidenceRound == agentmodel.MaxEvidenceRounds {
				return a.recoverRealtimeVoiceResponse(ctx, session, readState.toolCallIDs, readState.toolResults, emit)
			}
			vocabularyRequests, vocabularyDefinitions, err = mergeRealtimeVoiceVocabularyResolution(vocabularyCatalog, vocabularyRequests, vocabularyDefinitions, step.VocabularyRequests)
			if err != nil {
				return a.recoverRealtimeVoiceResponse(ctx, session, readState.toolCallIDs, readState.toolResults, emit)
			}
			continue
		}
		if step.Decision != agentmodel.InvestigationDecisionFinish {
			return a.recoverRealtimeVoiceResponse(ctx, session, readState.toolCallIDs, readState.toolResults, emit)
		}
		canonical, err := canonicalRealtimeVoiceInvestigationStep(canonicalIntent, step, observations, readEvidence)
		if err != nil {
			return a.recoverRealtimeVoiceResponse(ctx, session, readState.toolCallIDs, readState.toolResults, emit)
		}
		return a.completeRealtimeVoiceInvestigationOutcome(ctx, session, canonical.Intent, canonical.Resolutions, observations, readState.toolCallIDs, readState.toolResults, continueAfterClarification, emit)
	}
	return a.recoverRealtimeVoiceResponse(ctx, session, readState.toolCallIDs, readState.toolResults, emit)
}

func (a App) nextRealtimeVoiceInvestigation(ctx context.Context, session RealtimeVoiceSession, transcript string, conversationTurns []ports.AgentConversationTurn, investigation agentmodel.InvestigationInput, toolResults []ports.AgentToolResult, emit RealtimeVoiceEventSink) (agentmodel.InvestigationStep, error) {
	turn, err := session.languageInference.NextTurn(ctx, ports.LanguageInferenceInput{
		TenantID: session.TenantID, InventoryID: session.InventoryID, Principal: session.Principal,
		Transcript: transcript, ConversationTurns: safeRealtimeVoiceConversationTurns(conversationTurns),
		PromptTemplate: session.LanguagePromptTemplate,
		PreviousTurns:  investigation.EvidenceRound, Investigation: &investigation,
	})
	if err != nil {
		if diagnosticErr := emitRealtimeVoiceLanguageFailureDiagnostic(session, investigation, toolResults, realtimeVoiceFailureLanguageInference, err, emit); diagnosticErr != nil {
			return agentmodel.InvestigationStep{}, diagnosticErr
		}
		return agentmodel.InvestigationStep{}, realtimeVoiceProviderStageError{code: realtimeVoiceFailureLanguageInference, err: err}
	}
	if turn.Investigation == nil || turn.Investigation.Validate() != nil {
		return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
	}
	if err := emitRealtimeVoiceInvestigationDiagnostic(session, investigation, *turn.Investigation, emit); err != nil {
		return agentmodel.InvestigationStep{}, err
	}
	return *turn.Investigation, nil
}

func (a App) completeRealtimeVoiceInvestigationOutcome(ctx context.Context, session RealtimeVoiceSession, intent agentmodel.Intent, resolutions []agentmodel.Resolution, observations []agentmodel.CandidateObservation, toolCallIDs []string, toolResults []ports.AgentToolResult, continueAfterClarification bool, emit RealtimeVoiceEventSink) error {
	candidates := map[string]agentmodel.CandidateObservation{}
	for _, observation := range observations {
		candidates[observation.CandidateID] = observation
	}
	if intent.Kind != agentmodel.IntentKindChange || hasRealtimeVoiceInvestigationStatus(resolutions, agentmodel.ResolutionAmbiguous) ||
		hasRealtimeVoiceInvestigationStatus(resolutions, agentmodel.ResolutionUnsupported) || hasRealtimeVoiceInvestigationStatus(resolutions, agentmodel.ResolutionAbsent) {
		response, err := realtimeVoiceInvestigationResponse(intent, resolutions, candidates)
		if err != nil {
			return a.recoverRealtimeVoiceResponse(ctx, session, toolCallIDs, toolResults, emit)
		}
		return a.completeRealtimeVoiceResponse(ctx, session, response, toolCallIDs, toolResults, emit, continueAfterClarification)
	}
	if err := emitRealtimeVoiceProgress(session, realtimeVoiceProgressPlanning, "Preparing a safe plan.", emit); err != nil {
		return err
	}
	compiled, err := compileRealtimeVoiceActionPlan(intent, resolutions, candidates)
	if err != nil {
		return a.recoverRealtimeVoiceResponse(ctx, session, toolCallIDs, toolResults, emit)
	}
	if compiled.Disposition == realtimeVoicePlanNoOp {
		return a.completeRealtimeVoiceResponse(ctx, session, investigationResponse(ports.StructuredAgentResponseKindAnswer, compiled.NoOpSummary), toolCallIDs, toolResults, emit, continueAfterClarification)
	}
	record, err := a.CreateActionPlan(ctx, CreateActionPlanInput{
		Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID,
		Source: session.Source, RealtimeSessionID: session.ID, IntentSummary: compiled.IntentSummary,
		ModelInterpretationSummary: compiled.ModelInterpretationSummary, ConfirmationSummary: compiled.ConfirmationSummary,
		Commands: compiled.Commands, Risks: compiled.Risks,
	})
	if err != nil {
		return err
	}
	proposal, err := a.realtimeVoiceActionPlanProposal(ctx, session, record)
	if err != nil {
		return err
	}
	if err := emitRealtimeVoiceProgress(session, realtimeVoiceProgressReviewing, "Preparing a review.", emit); err != nil {
		return err
	}
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventActionPlanProposed, SessionID: session.ID, ActionPlan: &proposal})
}

func mergeRealtimeVoiceInvestigationObservations(existing, additions []agentmodel.CandidateObservation) []agentmodel.CandidateObservation {
	byKey := map[string]agentmodel.CandidateObservation{}
	order := []string{}
	for _, observation := range append(append([]agentmodel.CandidateObservation{}, existing...), additions...) {
		key := observation.ReferenceKey.String() + "\x00" + strings.TrimSpace(observation.CandidateID)
		if current, found := byKey[key]; found {
			observation = mergeRealtimeVoiceInvestigationObservation(current, observation)
		} else {
			order = append(order, key)
		}
		byKey[key] = observation
	}
	merged := make([]agentmodel.CandidateObservation, 0, len(order))
	for _, key := range order {
		merged = append(merged, byKey[key])
	}
	return merged
}
