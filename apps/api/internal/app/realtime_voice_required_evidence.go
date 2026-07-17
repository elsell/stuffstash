package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceRequiredEvidenceRequest(intent agentmodel.Intent, step agentmodel.InvestigationStep, observations []agentmodel.CandidateObservation, evidence []agentmodel.ReadEvidence) (agentmodel.SearchRequest, bool, error) {
	readKind, required := realtimeVoiceOperationRequiredRead(intent.Operation)
	if !required || step.Decision != agentmodel.InvestigationDecisionFinish {
		return agentmodel.SearchRequest{}, false, nil
	}
	subject, found := realtimeVoiceInvestigationResolution(step.Resolutions, agentmodel.SemanticReferenceSubject)
	if !found {
		return agentmodel.SearchRequest{}, false, nil
	}
	candidateID := ""
	if (subject.Status == agentmodel.ResolutionStrong || subject.Status == agentmodel.ResolutionPlausible) && len(subject.CandidateIDs) == 1 {
		candidateID = subject.CandidateIDs[0]
	} else {
		candidateID = realtimeVoiceRequiredEvidenceExactTarget(intent.SubjectMention, subject, observations)
	}
	if candidateID == "" {
		return agentmodel.SearchRequest{}, false, nil
	}
	for _, record := range evidence {
		if record.ReferenceKey == agentmodel.SemanticReferenceSubject && record.ReadKind == readKind && strings.TrimSpace(record.VisibleAssetID) == candidateID {
			return agentmodel.SearchRequest{}, false, nil
		}
	}
	var candidate agentmodel.CandidateObservation
	visible := false
	for _, observation := range observations {
		if observation.ReferenceKey == agentmodel.SemanticReferenceSubject && observation.CandidateID == candidateID {
			candidate = observation
			visible = true
			break
		}
	}
	if !visible {
		return agentmodel.SearchRequest{}, false, ports.ErrInvalidProviderInput
	}
	scope := agentmodel.LifecycleScopeActive
	if candidate.LifecycleState == string(agentmodel.LifecycleScopeArchived) {
		scope = agentmodel.LifecycleScopeArchived
	}
	request := agentmodel.SearchRequest{
		ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: readKind, Mention: intent.SubjectMention,
		KindHint: candidate.Kind, VisibleAssetID: candidateID, LifecycleScope: scope,
	}
	if request.Validate() != nil {
		return agentmodel.SearchRequest{}, false, ports.ErrInvalidProviderInput
	}
	return request, true, nil
}

func realtimeVoiceRequiredEvidenceExactTarget(mention string, resolution agentmodel.Resolution, observations []agentmodel.CandidateObservation) string {
	allowed := map[string]struct{}{}
	for _, candidateID := range resolution.CandidateIDs {
		allowed[candidateID] = struct{}{}
	}
	exact := ""
	for _, observation := range observations {
		if observation.ReferenceKey != agentmodel.SemanticReferenceSubject || normalizeRealtimeVoiceInvestigationTitle(observation.Title) != normalizeRealtimeVoiceInvestigationTitle(mention) {
			continue
		}
		if _, included := allowed[observation.CandidateID]; !included {
			continue
		}
		if exact != "" {
			return ""
		}
		exact = observation.CandidateID
	}
	return exact
}

func realtimeVoiceOperationRequiredRead(operation agentmodel.Operation) (agentmodel.InvestigationReadKind, bool) {
	switch operation {
	case agentmodel.OperationListContents:
		return agentmodel.InvestigationReadListContents, true
	case agentmodel.OperationDetail:
		return agentmodel.InvestigationReadAssetDetail, true
	case agentmodel.OperationAssetHistory:
		return agentmodel.InvestigationReadAssetHistory, true
	case agentmodel.OperationCheckoutHistory:
		return agentmodel.InvestigationReadCheckoutHistory, true
	default:
		return "", false
	}
}
