package app

import (
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type realtimeVoicePlanDisposition string

const (
	realtimeVoicePlanReady realtimeVoicePlanDisposition = "ready"
	realtimeVoicePlanNoOp  realtimeVoicePlanDisposition = "no_op"
)

type realtimeVoiceCompiledActionPlan struct {
	Disposition                realtimeVoicePlanDisposition
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	Commands                   []ActionPlanCommandInput
	Risks                      []string
	NoOpSummary                string
}

func compileRealtimeVoiceActionPlan(intent agentmodel.Intent, resolutions []agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) (realtimeVoiceCompiledActionPlan, error) {
	if intent.Validate() != nil || intent.Kind != agentmodel.IntentKindChange {
		return realtimeVoiceCompiledActionPlan{}, ports.ErrInvalidProviderInput
	}
	switch intent.Operation {
	case agentmodel.OperationCreate:
		return compileRealtimeVoiceCreatePlan(intent, resolutions, candidates)
	case agentmodel.OperationMove:
		return compileRealtimeVoiceMovePlan(intent, resolutions, candidates)
	case agentmodel.OperationArchive, agentmodel.OperationRestore, agentmodel.OperationCheckout, agentmodel.OperationReturn:
		return compileRealtimeVoiceSingleAssetPlan(intent, resolutions, candidates)
	default:
		return realtimeVoiceCompiledActionPlan{}, ports.ErrInvalidProviderInput
	}
}

func compileRealtimeVoiceCreatePlan(intent agentmodel.Intent, resolutions []agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) (realtimeVoiceCompiledActionPlan, error) {
	subject, ok := realtimeVoiceInvestigationResolution(resolutions, agentmodel.SemanticReferenceSubject)
	if !ok {
		return realtimeVoiceCompiledActionPlan{}, ports.ErrInvalidProviderInput
	}
	if subject.Status == agentmodel.ResolutionStrong || subject.Status == agentmodel.ResolutionPlausible {
		candidate, err := realtimeVoicePlanCandidateForResolution(subject, candidates)
		if err != nil {
			return realtimeVoiceCompiledActionPlan{}, err
		}
		return realtimeVoiceCompiledActionPlan{Disposition: realtimeVoicePlanNoOp, NoOpSummary: candidate.Title + " already exists in this inventory, so I did not prepare a duplicate."}, nil
	}
	if subject.Status != agentmodel.ResolutionMissing || len(subject.CandidateIDs) != 0 {
		return realtimeVoiceCompiledActionPlan{}, ports.ErrInvalidProviderInput
	}
	commands, parentAssetID, parentCommandID, createdDestinations, err := compileRealtimeVoiceDestinationPath(intent, resolutions, candidates)
	if err != nil {
		return realtimeVoiceCompiledActionPlan{}, err
	}
	arguments := map[string]any{"title": strings.TrimSpace(intent.SubjectMention), "kind": strings.TrimSpace(intent.NewAssetKind)}
	setRealtimeVoiceCompiledParent(arguments, parentAssetID, parentCommandID)
	commands = append(commands, ActionPlanCommandInput{ID: "create-subject", Kind: actionplan.CommandKindCreateAsset, Summary: "Create " + strings.TrimSpace(intent.SubjectMention), Arguments: arguments})
	destination := strings.Join(intent.DestinationPath, " / ")
	confirmation := "Create " + strings.TrimSpace(intent.SubjectMention)
	if destination != "" {
		confirmation += " in " + destination
	}
	confirmation += "?"
	compiled := realtimeVoiceCompiledActionPlan{
		Disposition: realtimeVoicePlanReady, IntentSummary: confirmation[:len(confirmation)-1],
		ModelInterpretationSummary: confirmation[:len(confirmation)-1], ConfirmationSummary: confirmation, Commands: commands,
	}
	if createdDestinations {
		compiled.Risks = []string{"This plan will create the missing destination path shown above."}
	}
	return compiled, nil
}

func compileRealtimeVoiceMovePlan(intent agentmodel.Intent, resolutions []agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) (realtimeVoiceCompiledActionPlan, error) {
	subject, ok := realtimeVoiceInvestigationResolution(resolutions, agentmodel.SemanticReferenceSubject)
	if !ok || (subject.Status != agentmodel.ResolutionStrong && subject.Status != agentmodel.ResolutionPlausible) {
		return realtimeVoiceCompiledActionPlan{}, ports.ErrInvalidProviderInput
	}
	item, err := realtimeVoicePlanCandidateForResolution(subject, candidates)
	if err != nil {
		return realtimeVoiceCompiledActionPlan{}, err
	}
	commands, parentAssetID, parentCommandID, createdDestinations, err := compileRealtimeVoiceDestinationPath(intent, resolutions, candidates)
	if err != nil || (parentAssetID == "" && parentCommandID == "") {
		if err != nil {
			return realtimeVoiceCompiledActionPlan{}, err
		}
		return realtimeVoiceCompiledActionPlan{}, ports.ErrInvalidProviderInput
	}
	arguments := map[string]any{"assetId": item.CandidateID}
	setRealtimeVoiceCompiledParent(arguments, parentAssetID, parentCommandID)
	commands = append(commands, ActionPlanCommandInput{ID: "move-subject", Kind: actionplan.CommandKindMoveAsset, Summary: "Move " + item.Title, Arguments: arguments})
	destination := strings.Join(intent.DestinationPath, " / ")
	confirmation := "Move " + item.Title + " to " + destination + "?"
	compiled := realtimeVoiceCompiledActionPlan{
		Disposition: realtimeVoicePlanReady, IntentSummary: "Move " + item.Title + " to " + destination,
		ModelInterpretationSummary: "Move the visible item to " + destination, ConfirmationSummary: confirmation, Commands: commands,
	}
	if createdDestinations {
		compiled.Risks = []string{"This plan will create the missing destination path shown above."}
	}
	return compiled, nil
}

func compileRealtimeVoiceDestinationPath(intent agentmodel.Intent, resolutions []agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) ([]ActionPlanCommandInput, string, string, bool, error) {
	commands := []ActionPlanCommandInput{}
	parentAssetID := ""
	parentCommandID := ""
	missingStarted := false
	for index, title := range intent.DestinationPath {
		key, _ := agentmodel.NewSemanticReferenceKey(fmt.Sprintf("destination.%d", index))
		resolution, ok := realtimeVoiceInvestigationResolution(resolutions, key)
		if !ok {
			return nil, "", "", false, ports.ErrInvalidProviderInput
		}
		if resolution.Status == agentmodel.ResolutionStrong || resolution.Status == agentmodel.ResolutionPlausible {
			if missingStarted {
				return nil, "", "", false, ports.ErrInvalidProviderInput
			}
			candidate, err := realtimeVoicePlanCandidateForResolution(resolution, candidates)
			if err != nil || (candidate.Kind != "location" && candidate.Kind != "container") || (index > 0 && candidate.ParentAssetID != parentAssetID) {
				return nil, "", "", false, ports.ErrInvalidProviderInput
			}
			parentAssetID = candidate.CandidateID
			parentCommandID = ""
			continue
		}
		if resolution.Status != agentmodel.ResolutionMissing || len(resolution.CandidateIDs) != 0 {
			return nil, "", "", false, ports.ErrInvalidProviderInput
		}
		missingStarted = true
		commandID := fmt.Sprintf("create-destination-%d", index)
		arguments := map[string]any{"title": strings.TrimSpace(title)}
		kind := actionplan.CommandKindCreateLocation
		if index > 0 {
			kind = actionplan.CommandKindCreateAsset
			arguments["kind"] = "container"
		}
		setRealtimeVoiceCompiledParent(arguments, parentAssetID, parentCommandID)
		commands = append(commands, ActionPlanCommandInput{ID: commandID, Kind: kind, Summary: "Create " + strings.TrimSpace(title), Arguments: arguments})
		parentAssetID = ""
		parentCommandID = commandID
	}
	return commands, parentAssetID, parentCommandID, missingStarted, nil
}

func compileRealtimeVoiceSingleAssetPlan(intent agentmodel.Intent, resolutions []agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) (realtimeVoiceCompiledActionPlan, error) {
	subject, ok := realtimeVoiceInvestigationResolution(resolutions, agentmodel.SemanticReferenceSubject)
	if !ok || (subject.Status != agentmodel.ResolutionStrong && subject.Status != agentmodel.ResolutionPlausible) {
		return realtimeVoiceCompiledActionPlan{}, ports.ErrInvalidProviderInput
	}
	candidate, err := realtimeVoicePlanCandidateForResolution(subject, candidates)
	if err != nil {
		return realtimeVoiceCompiledActionPlan{}, err
	}
	var commandKind actionplan.CommandKind
	var alreadySatisfied bool
	var noOp string
	switch intent.Operation {
	case agentmodel.OperationArchive:
		commandKind = actionplan.CommandKindArchiveAsset
		alreadySatisfied = candidate.LifecycleState == "archived"
		noOp = candidate.Title + " is already archived."
	case agentmodel.OperationRestore:
		commandKind = actionplan.CommandKindRestoreAsset
		alreadySatisfied = candidate.LifecycleState == "active"
		noOp = candidate.Title + " is already active."
	case agentmodel.OperationCheckout:
		commandKind = actionplan.CommandKindCheckoutAsset
		alreadySatisfied = candidate.CheckoutState == "checked_out"
		noOp = candidate.Title + " is already checked out."
	case agentmodel.OperationReturn:
		commandKind = actionplan.CommandKindReturnAsset
		alreadySatisfied = candidate.CheckoutState != "checked_out"
		noOp = candidate.Title + " is not checked out, so there is nothing to return."
	}
	if alreadySatisfied {
		return realtimeVoiceCompiledActionPlan{Disposition: realtimeVoicePlanNoOp, NoOpSummary: noOp}, nil
	}
	arguments := map[string]any{"assetId": candidate.CandidateID}
	if (intent.Operation == agentmodel.OperationCheckout || intent.Operation == agentmodel.OperationReturn) && strings.TrimSpace(intent.Details) != "" {
		arguments["details"] = strings.TrimSpace(intent.Details)
	}
	verb := strings.Title(string(intent.Operation))
	summary := verb + " " + candidate.Title
	return realtimeVoiceCompiledActionPlan{
		Disposition: realtimeVoicePlanReady, IntentSummary: summary, ModelInterpretationSummary: summary,
		ConfirmationSummary: summary + "?", Commands: []ActionPlanCommandInput{{ID: string(intent.Operation) + "-subject", Kind: commandKind, Summary: summary, Arguments: arguments}},
	}, nil
}

func realtimeVoicePlanCandidateForResolution(resolution agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) (agentmodel.CandidateObservation, error) {
	if len(resolution.CandidateIDs) != 1 {
		return agentmodel.CandidateObservation{}, ports.ErrInvalidProviderInput
	}
	candidate, exists := candidates[resolution.CandidateIDs[0]]
	if !exists || candidate.ReferenceKey != resolution.ReferenceKey {
		return agentmodel.CandidateObservation{}, ports.ErrInvalidProviderInput
	}
	return candidate, nil
}

func setRealtimeVoiceCompiledParent(arguments map[string]any, parentAssetID, parentCommandID string) {
	if parentCommandID != "" {
		arguments["parentCommandId"] = parentCommandID
	} else if parentAssetID != "" {
		arguments["parentAssetId"] = parentAssetID
	}
}
