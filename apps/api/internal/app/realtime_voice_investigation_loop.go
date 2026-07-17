package app

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeVoiceInvestigationVersion = "voice-investigation-v2"

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
	if step.Intent.Operation == agentmodel.OperationUnsupported {
		response, responseErr := realtimeVoiceInvestigationResponse(step.Intent, nil, nil)
		if responseErr != nil {
			return a.recoverRealtimeVoiceResponse(ctx, session, nil, nil, emit)
		}
		return a.completeRealtimeVoiceResponse(ctx, session, response, nil, nil, emit, continueAfterClarification)
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
	if canonicalIntent.Validate() != nil {
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
			if evidenceRound == agentmodel.MaxEvidenceRounds || step.Decision != agentmodel.InvestigationDecisionSearchAgain ||
				!realtimeVoiceDestinationRepairAllowed(transcript, canonicalIntent, step.Intent) ||
				!realtimeVoiceDestinationRepairRequestsValid(step.Intent, step.SearchRequests) {
				return a.recoverRealtimeVoiceResponse(ctx, session, readState.toolCallIDs, readState.toolResults, emit)
			}
			canonicalIntent = step.Intent
			step.SearchRequests = realtimeVoiceDestinationRequests(step.SearchRequests)
			requests = realtimeVoiceSubjectRequests(requests)
			observations = realtimeVoiceSubjectObservations(observations)
			readEvidence = realtimeVoiceSubjectReadEvidence(readEvidence)
			readState.resetDestinationScope()
			continue
		}
		requiredRead, required, requiredErr := realtimeVoiceRequiredEvidenceRequest(canonicalIntent, step, observations, readEvidence)
		if requiredErr != nil {
			return a.recoverRealtimeVoiceResponse(ctx, session, readState.toolCallIDs, readState.toolResults, emit)
		}
		if required {
			step = agentmodel.InvestigationStep{
				Decision: agentmodel.InvestigationDecisionSearchAgain, Intent: canonicalIntent,
				SearchRequests: []agentmodel.SearchRequest{requiredRead}, Rationale: "Complete operation-required typed evidence.",
			}
		}
		if completed, ok := realtimeVoiceExactOrZeroCompletion(canonicalIntent, step, observations, readEvidence); ok {
			step = completed
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

func realtimeVoiceDestinationRepairAllowed(transcript string, original, repaired agentmodel.Intent) bool {
	if (original.Operation != agentmodel.OperationCreate && original.Operation != agentmodel.OperationMove) ||
		original.RequestShape != repaired.RequestShape || original.Kind != repaired.Kind || original.Operation != repaired.Operation ||
		normalizeRealtimeVoiceSemanticMention(original.SubjectMention) != normalizeRealtimeVoiceSemanticMention(repaired.SubjectMention) ||
		strings.TrimSpace(original.NewAssetKind) != strings.TrimSpace(repaired.NewAssetKind) || strings.TrimSpace(original.Details) != strings.TrimSpace(repaired.Details) ||
		len(repaired.DestinationPath) < len(original.DestinationPath) || repaired.Validate() != nil {
		return false
	}
	originalSegments := map[string]int{}
	for _, mention := range original.DestinationPath {
		key := normalizeRealtimeVoiceSemanticMention(mention)
		originalSegments[key]++
	}
	repairedSegments := map[string]int{}
	normalizedTranscript := normalizeRealtimeVoiceInvestigationTitle(transcript)
	normalizedSubject := normalizeRealtimeVoiceSemanticMention(repaired.SubjectMention)
	for index, mention := range repaired.DestinationPath {
		normalizedMention := normalizeRealtimeVoiceInvestigationTitle(mention)
		semanticMention := normalizeRealtimeVoiceSemanticMention(mention)
		if semanticMention == "" || semanticMention == normalizedSubject ||
			(!containsRealtimeVoiceTokenPhrase(normalizedTranscript, normalizedMention) && !containsRealtimeVoiceTokenPhrase(normalizedTranscript, semanticMention)) {
			return false
		}
		if index < len(original.DestinationPath) && semanticMention == normalizeRealtimeVoiceSemanticMention(original.DestinationPath[index]) && repaired.DestinationKinds[index] != original.DestinationKinds[index] {
			return false
		}
		key := semanticMention
		repairedSegments[key]++
	}
	for key, count := range originalSegments {
		if repairedSegments[key] < count {
			return false
		}
	}
	return true
}

func containsRealtimeVoiceTokenPhrase(normalizedText, normalizedPhrase string) bool {
	if normalizedText == "" || normalizedPhrase == "" {
		return false
	}
	return strings.Contains(" "+normalizedText+" ", " "+normalizedPhrase+" ")
}

func realtimeVoiceDestinationRepairRequestsValid(intent agentmodel.Intent, requests []agentmodel.SearchRequest) bool {
	if len(requests) < len(intent.DestinationPath) || len(requests) > len(intent.DestinationPath)+1 {
		return false
	}
	seen := map[agentmodel.SemanticReferenceKey]bool{}
	subjectSeen := false
	for _, request := range requests {
		if request.ReferenceKey == agentmodel.SemanticReferenceSubject {
			if subjectSeen || normalizeRealtimeVoiceSemanticMention(request.Mention) != normalizeRealtimeVoiceSemanticMention(intent.SubjectMention) {
				return false
			}
			subjectSeen = true
			continue
		}
		if request.ReadKind != agentmodel.InvestigationReadSearchAssets || strings.TrimSpace(request.VisibleAssetID) != "" {
			return false
		}
		var index int
		if _, err := fmt.Sscanf(request.ReferenceKey.String(), "destination.%d", &index); err != nil || index < 0 || index >= len(intent.DestinationPath) || seen[request.ReferenceKey] ||
			normalizeRealtimeVoiceSemanticMention(request.Mention) != normalizeRealtimeVoiceSemanticMention(intent.DestinationPath[index]) || request.KindHint != string(intent.DestinationKinds[index]) {
			return false
		}
		seen[request.ReferenceKey] = true
	}
	return len(seen) == len(intent.DestinationPath)
}

func normalizeRealtimeVoiceSemanticMention(value string) string {
	words := strings.Fields(normalizeRealtimeVoiceInvestigationTitle(value))
	if len(words) > 1 {
		switch words[0] {
		case "my", "the", "a", "an":
			words = words[1:]
		}
	}
	return strings.Join(words, " ")
}

func realtimeVoiceDestinationRequests(requests []agentmodel.SearchRequest) []agentmodel.SearchRequest {
	return slices.DeleteFunc(append([]agentmodel.SearchRequest{}, requests...), func(request agentmodel.SearchRequest) bool {
		return request.ReferenceKey == agentmodel.SemanticReferenceSubject
	})
}

func realtimeVoiceSubjectRequests(requests []agentmodel.SearchRequest) []agentmodel.SearchRequest {
	return slices.DeleteFunc(append([]agentmodel.SearchRequest{}, requests...), func(request agentmodel.SearchRequest) bool {
		return request.ReferenceKey != agentmodel.SemanticReferenceSubject
	})
}

func realtimeVoiceSubjectObservations(observations []agentmodel.CandidateObservation) []agentmodel.CandidateObservation {
	return slices.DeleteFunc(append([]agentmodel.CandidateObservation{}, observations...), func(observation agentmodel.CandidateObservation) bool {
		return observation.ReferenceKey != agentmodel.SemanticReferenceSubject
	})
}

func realtimeVoiceSubjectReadEvidence(evidence []agentmodel.ReadEvidence) []agentmodel.ReadEvidence {
	return slices.DeleteFunc(append([]agentmodel.ReadEvidence{}, evidence...), func(record agentmodel.ReadEvidence) bool {
		return record.ReferenceKey != agentmodel.SemanticReferenceSubject
	})
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
	if turn.Investigation == nil || turn.Investigation.Intent.RequestShape == "" {
		return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
	}
	turn.Investigation.Intent = agentmodel.CanonicalizeIntent(turn.Investigation.Intent)
	if turn.Investigation.Validate() != nil {
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
		hasRealtimeVoiceInvestigationStatus(resolutions, agentmodel.ResolutionUnsupported) || hasRealtimeVoiceInvestigationStatus(resolutions, agentmodel.ResolutionAbsent) ||
		hasRealtimeVoicePlausibleDestination(resolutions) {
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
