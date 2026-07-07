package app

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) WithRealtimeVoiceProviders(stt ports.SpeechToTextProvider, lm ports.LanguageInferenceProvider, tts ports.TextToSpeechProvider) App {
	a.speechToText = stt
	a.languageInference = lm
	a.textToSpeech = tts
	a.realtimeVoiceProviders = staticRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{
		SpeechToText:      stt,
		LanguageInference: lm,
		TextToSpeech:      tts,
	}}
	return a
}

func (a App) StartRealtimeVoiceSession(ctx context.Context, input RealtimeVoiceSessionInput) (RealtimeVoiceSession, error) {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return RealtimeVoiceSession{}, err
	}
	if input.Source != RealtimeVoiceSourceMobile {
		return RealtimeVoiceSession{}, apperrors.ErrInvalidInput
	}
	if input.InputAudio.MimeType != "audio/mp4" || input.InputAudio.Channels != 1 {
		return RealtimeVoiceSession{}, apperrors.ErrInvalidInput
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return RealtimeVoiceSession{}, err
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return RealtimeVoiceSession{}, err
	}
	providers, err := a.realtimeVoiceProviders.ResolveRealtimeVoiceProviders(ctx, ports.RealtimeVoiceProviderResolutionInput{
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Principal:   input.Principal,
	})
	if err != nil {
		return RealtimeVoiceSession{}, err
	}
	if providers.SpeechToText == nil || providers.LanguageInference == nil || providers.TextToSpeech == nil {
		return RealtimeVoiceSession{}, apperrors.ErrInvalidInput
	}

	sessionID := a.newRealtimeVoiceID()
	if strings.TrimSpace(sessionID) == "" {
		return RealtimeVoiceSession{}, apperrors.ErrInvalidInput
	}
	session := RealtimeVoiceSession{
		ID:                         sessionID,
		TenantID:                   input.TenantID,
		InventoryID:                input.InventoryID,
		Principal:                  input.Principal,
		Source:                     input.Source,
		InputAudio:                 input.InputAudio,
		OutputAudio:                input.OutputAudio,
		SpeechToTextProfileID:      providers.SpeechToTextProfileID,
		LanguageInferenceProfileID: providers.LanguageInferenceProfileID,
		TextToSpeechProfileID:      providers.TextToSpeechProfileID,
		LanguagePromptTemplate:     providers.LanguagePromptTemplate,
		DeveloperDiagnostics:       input.DeveloperDiagnostics,
		speechToText:               providers.SpeechToText,
		languageInference:          providers.LanguageInference,
		textToSpeech:               providers.TextToSpeech,
	}
	now := a.clock.Now()
	if err := a.realtimeSessions.SaveRealtimeSession(ctx, ports.RealtimeSessionRecord{
		ID:                         session.ID,
		TenantID:                   session.TenantID,
		InventoryID:                session.InventoryID,
		PrincipalID:                session.Principal.ID,
		Source:                     session.Source,
		State:                      ports.RealtimeSessionStateStarted,
		SpeechToTextProfileID:      session.SpeechToTextProfileID,
		LanguageInferenceProfileID: session.LanguageInferenceProfileID,
		TextToSpeechProfileID:      session.TextToSpeechProfileID,
		StartedAt:                  now,
		LastActivityAt:             now,
	}); err != nil {
		return RealtimeVoiceSession{}, err
	}
	return session, nil
}

func (a App) RunRealtimeVoiceQuery(ctx context.Context, input RealtimeVoiceQueryInput, emit RealtimeVoiceEventSink) (err error) {
	defer func() {
		if err != nil && strings.TrimSpace(input.Session.ID) != "" {
			if errors.Is(err, context.Canceled) {
				_ = a.markRealtimeVoiceSessionOutcome(context.Background(), input.Session, ports.RealtimeSessionStateCancelled, "")
				return
			}
			safeCode := realtimeVoiceErrorCode(err)
			if a.observer != nil {
				a.observer.Record(ctx, ports.Event{
					Name:    ports.EventRealtimeVoiceFailed,
					Message: "realtime voice failed safely",
					Fields: map[string]string{
						"tenant_id":         input.Session.TenantID.String(),
						"inventory_id":      input.Session.InventoryID.String(),
						"principal_id":      input.Session.Principal.ID.String(),
						"session_id":        input.Session.ID,
						"safe_failure_code": safeCode,
						"error":             safeRealtimeVoiceErrorDetail(err),
					},
				})
			}
			_ = a.markRealtimeVoiceSessionOutcome(ctx, input.Session, ports.RealtimeSessionStateFailed, safeCode)
		}
	}()
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	if len(input.AudioChunks) == 0 {
		return ports.ErrInvalidProviderInput
	}

	if input.Session.speechToText == nil || input.Session.languageInference == nil || input.Session.textToSpeech == nil {
		return apperrors.ErrInvalidInput
	}
	transcription, err := input.Session.speechToText.Transcribe(ctx, ports.SpeechToTextInput{
		TenantID:    input.Session.TenantID,
		InventoryID: input.Session.InventoryID,
		Principal:   input.Session.Principal,
		AudioFormat: input.Session.InputAudio,
		AudioChunks: input.AudioChunks,
	})
	if err != nil {
		return realtimeVoiceProviderStageError{code: realtimeVoiceFailureSpeechToText, err: err}
	}
	transcript := strings.TrimSpace(transcription.Transcript)
	if transcript == "" {
		return ports.ErrInvalidProviderInput
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTranscriptFinal, SessionID: input.Session.ID, Text: transcript}); err != nil {
		return err
	}
	effectiveTranscript := realtimeVoiceEffectiveTranscript(transcript, input.ConversationTurns)
	if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressUnderstanding, "Understanding your request.", emit); err != nil {
		return err
	}
	if response, ok := realtimeVoiceUnsafeUnsupportedTranscriptResponse(effectiveTranscript); ok {
		return a.completeRealtimeVoiceResponse(ctx, input.Session, response, nil, nil, emit, input.ContinueAfterClarification)
	}
	if response, ok := realtimeVoiceAmbiguousDestinationTranscriptResponse(effectiveTranscript); ok {
		return a.completeRealtimeVoiceResponse(ctx, input.Session, response, nil, nil, emit, input.ContinueAfterClarification)
	}

	toolResults := []ports.AgentToolResult{}
	toolCallIDs := []string{}
	executedToolCalls := map[string]struct{}{}
	visibleAssetIDs := map[string]struct{}{}
	for turn := 0; turn < realtimeVoiceToolTurnBudget; turn++ {
		if call, diagnosticTitle, ok := realtimeVoiceServerSelectedReadCallWithoutModel(effectiveTranscript, turn, toolResults, ""); ok {
			if strings.TrimSpace(call.ID) == "" {
				call.ID = a.newRealtimeVoiceID()
			}
			if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressExploring, "Checking your inventory.", emit); err != nil {
				return err
			}
			proposal, err := a.executeRealtimeVoiceServerSelectedRead(ctx, input.Session, effectiveTranscript, call, diagnosticTitle, &toolResults, &toolCallIDs, executedToolCalls, visibleAssetIDs, emit)
			if err != nil {
				return err
			}
			if proposal != nil {
				if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressReviewing, "Preparing a review.", emit); err != nil {
					return err
				}
				if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventActionPlanProposed, SessionID: input.Session.ID, ActionPlan: proposal}); err != nil {
					return err
				}
				return nil
			}
			if response, ok := realtimeVoiceMissingMoveSourceResponse(effectiveTranscript, toolResults); ok {
				return a.completeRealtimeVoiceResponse(ctx, input.Session, response, toolCallIDs, toolResults, emit, input.ContinueAfterClarification)
			}
			if realtimeVoiceShouldFinalizeReadOnlyAfterToolTurn(effectiveTranscript, toolResults) {
				return a.finalizeRealtimeVoiceWithToolResults(ctx, input.Session, effectiveTranscript, input.ConversationTurns, toolResults, toolCallIDs, turn+1, emit, input.ContinueAfterClarification)
			}
			continue
		}
		tools := realtimeVoiceToolDescriptors()
		planOnly := realtimeVoiceShouldUseConstrainedPlanner(effectiveTranscript, turn, toolResults)
		requireToolCall := !planOnly && realtimeVoiceShouldRequireReadTool(effectiveTranscript, turn, toolResults)
		if requireToolCall {
			tools = realtimeVoiceReadToolsForTurn(effectiveTranscript, turn, toolResults)
		}
		if planOnly {
			if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressPlanning, "Preparing a safe plan.", emit); err != nil {
				return err
			}
		} else if requireToolCall {
			if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressExploring, "Checking your inventory.", emit); err != nil {
				return err
			}
		}
		modelTurn, err := input.Session.languageInference.NextTurn(ctx, ports.LanguageInferenceInput{
			TenantID:           input.Session.TenantID,
			InventoryID:        input.Session.InventoryID,
			Principal:          input.Session.Principal,
			Transcript:         effectiveTranscript,
			ConversationTurns:  safeRealtimeVoiceConversationTurns(input.ConversationTurns),
			PromptTemplate:     input.Session.LanguagePromptTemplate,
			Tools:              tools,
			ToolResults:        toolResults,
			PreviousTurns:      turn,
			PlanOnly:           planOnly,
			RequireToolCall:    requireToolCall,
			IncludeDiagnostics: input.Session.DeveloperDiagnostics,
		})
		if err != nil {
			if diagnosticErr := emitRealtimeVoiceLanguageFailureDiagnostic(input.Session, turn+1, false, toolResults, realtimeVoiceFailureLanguageInference, err, emit); diagnosticErr != nil {
				return diagnosticErr
			}
			return realtimeVoiceProviderStageError{code: realtimeVoiceFailureLanguageInference, err: err}
		}
		if err := emitRealtimeVoiceDiagnostics(input.Session, modelTurn.Diagnostics, emit); err != nil {
			return err
		}
		if input.Session.DeveloperDiagnostics {
			for _, call := range modelTurn.ToolCalls {
				if err := emitRealtimeVoiceDiagnostic(input.Session.ID, "Tool call requested", realtimeVoiceToolCallDiagnosticDetail(call), emit); err != nil {
					return err
				}
			}
		}
		if modelTurn.Final != nil {
			if call, diagnosticTitle, ok := realtimeVoiceServerSelectedReadBeforeFinalCall(effectiveTranscript, toolResults); ok {
				call.ID = a.newRealtimeVoiceID()
				if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressExploring, "Checking your inventory.", emit); err != nil {
					return err
				}
				proposal, err := a.executeRealtimeVoiceServerSelectedRead(ctx, input.Session, effectiveTranscript, call, diagnosticTitle, &toolResults, &toolCallIDs, executedToolCalls, visibleAssetIDs, emit)
				if err != nil {
					return err
				}
				if proposal != nil {
					if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressReviewing, "Preparing a review.", emit); err != nil {
						return err
					}
					if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventActionPlanProposed, SessionID: input.Session.ID, ActionPlan: proposal}); err != nil {
						return err
					}
					return nil
				}
				continue
			}
			if realtimeVoiceShouldRepairCreateClarification(effectiveTranscript, *modelTurn.Final, toolResults) {
				result, resultErr := realtimeVoiceFinalClarificationRepairResult(a.newRealtimeVoiceID())
				if resultErr != nil {
					return resultErr
				}
				toolResults = append(toolResults, result)
				if input.Session.DeveloperDiagnostics {
					if err := emitRealtimeVoiceDiagnostic(input.Session.ID, "Final clarification repaired", realtimeVoiceToolResultDiagnosticDetail(result), emit); err != nil {
						return err
					}
				}
				continue
			}
			if realtimeVoiceShouldRepairWriteClaimAfterFailedProposal(effectiveTranscript, *modelTurn.Final, toolResults) {
				result, resultErr := realtimeVoiceFinalWriteClaimRepairResult(a.newRealtimeVoiceID())
				if resultErr != nil {
					return resultErr
				}
				toolResults = append(toolResults, result)
				if input.Session.DeveloperDiagnostics {
					if err := emitRealtimeVoiceDiagnostic(input.Session.ID, "Final write claim repaired", realtimeVoiceToolResultDiagnosticDetail(result), emit); err != nil {
						return err
					}
				}
				continue
			}
			if err := validateRealtimeVoiceFinalResponse(*modelTurn.Final); err != nil {
				if recoverableRealtimeVoiceToolError(err) {
					return a.recoverRealtimeVoiceResponse(ctx, input.Session, toolCallIDs, toolResults, emit)
				}
				return err
			}
			if input.Session.DeveloperDiagnostics {
				if err := emitRealtimeVoiceDiagnostic(input.Session.ID, "Structured final response", realtimeVoiceFinalDiagnosticDetail(*modelTurn.Final), emit); err != nil {
					return err
				}
			}
			return a.completeRealtimeVoiceResponse(ctx, input.Session, *modelTurn.Final, toolCallIDs, toolResults, emit, input.ContinueAfterClarification)
		}
		if len(modelTurn.ToolCalls) == 0 {
			return a.recoverRealtimeVoiceResponse(ctx, input.Session, toolCallIDs, toolResults, emit)
		}
		for _, call := range modelTurn.ToolCalls {
			if selectedCall, diagnosticTitle := realtimeVoiceServerSelectedReadCall(effectiveTranscript, turn, toolResults, call); diagnosticTitle != "" {
				call = selectedCall
				if input.Session.DeveloperDiagnostics {
					if err := emitRealtimeVoiceDiagnostic(input.Session.ID, diagnosticTitle, realtimeVoiceToolCallDiagnosticDetail(call), emit); err != nil {
						return err
					}
				}
			}
			signature, err := realtimeVoiceToolCallSignature(call)
			if err != nil {
				return err
			}
			if _, duplicate := executedToolCalls[signature]; duplicate {
				toolCallID := strings.TrimSpace(call.ID)
				if toolCallID == "" {
					toolCallID = a.newRealtimeVoiceID()
				}
				toolLabel := realtimeVoiceToolLabel(call.Name)
				toolCallIDs = append(toolCallIDs, toolCallID)
				if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallFailed, SessionID: input.Session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Code: "duplicate_tool_request", Message: "I already checked that."}); err != nil {
					return err
				}
				result, resultErr := realtimeVoiceToolErrorResult(ports.AgentToolCall{
					ID:        toolCallID,
					Name:      call.Name,
					Arguments: call.Arguments,
				}, "duplicate_tool_request", "This exact tool request has already been executed. Use the existing tool result, make a distinct tool request, ask for clarification, or produce a final response.", true)
				if resultErr != nil {
					return resultErr
				}
				toolResults = append(toolResults, result)
				continue
			}
			toolCallID := strings.TrimSpace(call.ID)
			if toolCallID == "" {
				toolCallID = a.newRealtimeVoiceID()
			}
			toolLabel := realtimeVoiceToolLabel(call.Name)
			toolCallIDs = append(toolCallIDs, toolCallID)
			if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallStarted, SessionID: input.Session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Status: "searching"}); err != nil {
				return err
			}
			executableCall := ports.AgentToolCall{
				ID:        toolCallID,
				Name:      call.Name,
				Arguments: call.Arguments,
			}
			result, proposal, err := a.executeRealtimeVoiceTool(ctx, input.Session, effectiveTranscript, toolResults, executableCall, visibleAssetIDs)
			if err != nil {
				if !recoverableRealtimeVoiceToolError(err) {
					_ = emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallFailed, SessionID: input.Session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Code: "tool_failed", Message: "I could not check that safely."})
					return err
				}
				if input.Session.DeveloperDiagnostics {
					if diagnosticErr := emitRealtimeVoiceDiagnostic(input.Session.ID, "Tool validation failed", safeRealtimeVoiceErrorDetail(err), emit); diagnosticErr != nil {
						return diagnosticErr
					}
				}
				if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallFailed, SessionID: input.Session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Code: "invalid_tool_request", Message: "I need a little more detail to do that safely."}); err != nil {
					return err
				}
				result, resultErr := realtimeVoiceToolErrorResult(executableCall, "invalid_tool_request", realtimeVoiceInvalidToolRequestRepairMessage(executableCall.Name), true)
				if resultErr != nil {
					return resultErr
				}
				if input.Session.DeveloperDiagnostics {
					if err := emitRealtimeVoiceDiagnostic(input.Session.ID, "Tool result received", realtimeVoiceToolResultDiagnosticDetail(result), emit); err != nil {
						return err
					}
				}
				executedToolCalls[signature] = struct{}{}
				toolResults = append(toolResults, result)
				continue
			}
			executedToolCalls[signature] = struct{}{}
			if realtimeVoiceToolReturnsVisibleAssetItems(call.Name) {
				if err := collectRealtimeVoiceVisibleAssetIDs(result, visibleAssetIDs); err != nil {
					return err
				}
			}
			toolResults = append(toolResults, result)
			if input.Session.DeveloperDiagnostics {
				if err := emitRealtimeVoiceDiagnostic(input.Session.ID, "Tool result received", realtimeVoiceToolResultDiagnosticDetail(result), emit); err != nil {
					return err
				}
			}
			if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallCompleted, SessionID: input.Session.ID, ToolCallID: toolCallID, ToolLabel: toolLabel, Status: realtimeVoiceToolCompletionStatus(result)}); err != nil {
				return err
			}
			if proposal != nil {
				if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressReviewing, "Preparing a review.", emit); err != nil {
					return err
				}
				if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventActionPlanProposed, SessionID: input.Session.ID, ActionPlan: proposal}); err != nil {
					return err
				}
				return nil
			}
		}
		if call, diagnosticTitle, ok := realtimeVoiceServerSelectedExplorationCall(effectiveTranscript, turn, toolResults, a.newRealtimeVoiceID()); ok {
			if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressExploring, "Checking one more match.", emit); err != nil {
				return err
			}
			proposal, err := a.executeRealtimeVoiceServerSelectedRead(ctx, input.Session, effectiveTranscript, call, diagnosticTitle, &toolResults, &toolCallIDs, executedToolCalls, visibleAssetIDs, emit)
			if err != nil {
				return err
			}
			if proposal != nil {
				if err := emitRealtimeVoiceProgress(input.Session, realtimeVoiceProgressReviewing, "Preparing a review.", emit); err != nil {
					return err
				}
				if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventActionPlanProposed, SessionID: input.Session.ID, ActionPlan: proposal}); err != nil {
					return err
				}
				return nil
			}
		}
		if realtimeVoiceShouldFinalizeReadOnlyAfterToolTurn(effectiveTranscript, toolResults) {
			return a.finalizeRealtimeVoiceWithToolResults(ctx, input.Session, effectiveTranscript, input.ConversationTurns, toolResults, toolCallIDs, turn+1, emit, input.ContinueAfterClarification)
		}
		if response, ok := realtimeVoiceMissingMoveSourceResponse(effectiveTranscript, toolResults); ok {
			return a.completeRealtimeVoiceResponse(ctx, input.Session, response, toolCallIDs, toolResults, emit, input.ContinueAfterClarification)
		}
	}
	return a.finalizeRealtimeVoiceAfterToolBudget(ctx, input.Session, effectiveTranscript, input.ConversationTurns, toolResults, toolCallIDs, emit, input.ContinueAfterClarification)
}

func (a App) finalizeRealtimeVoiceAfterToolBudget(ctx context.Context, session RealtimeVoiceSession, transcript string, conversationTurns []ports.AgentConversationTurn, toolResults []ports.AgentToolResult, toolCallIDs []string, emit RealtimeVoiceEventSink, continueAfterClarification ...bool) error {
	return a.finalizeRealtimeVoiceWithToolResults(ctx, session, transcript, conversationTurns, toolResults, toolCallIDs, realtimeVoiceToolTurnBudget, emit, continueAfterClarification...)
}

func (a App) finalizeRealtimeVoiceWithToolResults(ctx context.Context, session RealtimeVoiceSession, transcript string, conversationTurns []ports.AgentConversationTurn, toolResults []ports.AgentToolResult, toolCallIDs []string, previousTurns int, emit RealtimeVoiceEventSink, continueAfterClarification ...bool) error {
	if err := emitRealtimeVoiceProgress(session, realtimeVoiceProgressAnswering, "Preparing an answer.", emit); err != nil {
		return err
	}
	modelTurn, err := session.languageInference.NextTurn(ctx, ports.LanguageInferenceInput{
		TenantID:           session.TenantID,
		InventoryID:        session.InventoryID,
		Principal:          session.Principal,
		Transcript:         transcript,
		ConversationTurns:  safeRealtimeVoiceConversationTurns(conversationTurns),
		PromptTemplate:     session.LanguagePromptTemplate,
		ToolResults:        toolResults,
		PreviousTurns:      previousTurns,
		FinalOnly:          true,
		IncludeDiagnostics: session.DeveloperDiagnostics,
	})
	if err != nil {
		if diagnosticErr := emitRealtimeVoiceLanguageFailureDiagnostic(session, previousTurns+1, true, toolResults, realtimeVoiceFailureLanguageInference, err, emit); diagnosticErr != nil {
			return diagnosticErr
		}
		return realtimeVoiceProviderStageError{code: realtimeVoiceFailureLanguageInference, err: err}
	}
	if err := emitRealtimeVoiceDiagnostics(session, modelTurn.Diagnostics, emit); err != nil {
		return err
	}
	if modelTurn.Final == nil {
		return a.recoverRealtimeVoiceResponse(ctx, session, toolCallIDs, toolResults, emit)
	}
	if err := validateRealtimeVoiceFinalResponse(*modelTurn.Final); err != nil {
		if recoverableRealtimeVoiceToolError(err) {
			return a.recoverRealtimeVoiceResponse(ctx, session, toolCallIDs, toolResults, emit)
		}
		return err
	}
	if session.DeveloperDiagnostics {
		if err := emitRealtimeVoiceDiagnostic(session.ID, "Structured final response", realtimeVoiceFinalDiagnosticDetail(*modelTurn.Final), emit); err != nil {
			return err
		}
	}
	return a.completeRealtimeVoiceResponse(ctx, session, *modelTurn.Final, toolCallIDs, toolResults, emit, continueAfterClarification...)
}

func realtimeVoiceToolCallSignature(call ports.AgentToolCall) (string, error) {
	payload, err := json.Marshal(call.Arguments)
	if err != nil {
		return "", ports.ErrInvalidProviderInput
	}
	return strings.TrimSpace(call.Name) + ":" + string(payload), nil
}

func realtimeVoiceInvalidToolRequestRepairMessage(toolName string) string {
	if strings.TrimSpace(toolName) == RealtimeVoiceToolProposeActionPlan {
		return "The action-plan request was invalid or incomplete. Retry with corrected structured arguments. For existing assets, assetId and parentAssetId must be opaque assetId values copied exactly from successful read tool results; never use titles or guessed IDs. For a new item, use one create_asset command with title or name and kind item; never include assetId and never add a move_asset command for that newly-created item. Put the new item directly in an existing visible parent with parentAssetId, or in a newly-created parent with parentCommandId. For missing destinations, create every missing location/container first, then reference those create commands with parentCommandId. If a missing container belongs inside an existing visible location, create the container with parentAssetId set to that visible location assetId, then create or move the requested item into the container with parentCommandId."
	}
	return "The tool request was invalid or incomplete. Retry with corrected, authorized, structured arguments, or ask the user for clarification."
}

func realtimeVoiceToolCallDiagnosticDetail(call ports.AgentToolCall) string {
	payload, err := json.MarshalIndent(map[string]any{
		"name":      call.Name,
		"arguments": redactRealtimeVoiceDiagnosticValue(call.Arguments),
	}, "", "  ")
	if err != nil {
		return "Tool call arguments could not be rendered safely."
	}
	return safeRealtimeVoiceDiagnosticText(string(payload), 4000)
}

func emitRealtimeVoiceDiagnostics(session RealtimeVoiceSession, diagnostics []ports.LanguageInferenceDiagnostic, emit RealtimeVoiceEventSink) error {
	if !session.DeveloperDiagnostics {
		return nil
	}
	for _, diagnostic := range diagnostics {
		if err := emitRealtimeVoiceDiagnostic(session.ID, diagnostic.Title, diagnostic.Detail, emit); err != nil {
			return err
		}
	}
	return nil
}

func emitRealtimeVoiceDiagnostic(sessionID string, title string, detail string, emit RealtimeVoiceEventSink) error {
	message := safeRealtimeVoiceDiagnosticText(title, 120)
	if message == "" {
		message = "Agent diagnostic"
	}
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAgentDiagnostic, SessionID: sessionID, Message: message, Detail: safeRealtimeVoiceDiagnosticText(detail, 4000)})
}

func realtimeVoiceFinalDiagnosticDetail(response ports.StructuredAgentResponse) string {
	payload, err := json.MarshalIndent(map[string]any{
		"kind":            response.Kind,
		"spokenResponse":  response.SpokenResponse,
		"displayResponse": response.DisplayResponse,
	}, "", "  ")
	if err != nil {
		return "Final response could not be rendered safely."
	}
	return safeRealtimeVoiceDiagnosticText(string(payload), 4000)
}

func realtimeVoiceToolResultDiagnosticDetail(result ports.AgentToolResult) string {
	payload, err := json.MarshalIndent(map[string]any{
		"name":    result.Name,
		"content": redactRealtimeVoiceDiagnosticString(result.Content),
	}, "", "  ")
	if err != nil {
		return "Tool result could not be rendered safely."
	}
	return safeRealtimeVoiceDiagnosticText(string(payload), 4000)
}

func safeRealtimeVoiceDiagnosticText(value string, maxLength int) string {
	trimmed := strings.TrimSpace(redactRealtimeVoiceDiagnosticString(value))
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= maxLength {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:maxLength]) + " ..."
}

func redactRealtimeVoiceDiagnosticString(value string) string {
	value = realtimeVoiceDiagnosticBearerPattern.ReplaceAllString(value, "[redacted-bearer] [redacted]")
	value = realtimeVoiceDiagnosticAssignmentPattern.ReplaceAllString(value, "$1[redacted]")
	replacer := strings.NewReplacer(
		"apiKey", "[redacted-key]",
		"api_key", "[redacted-key]",
		"authorization", "[redacted-authorization]",
		"credential", "[redacted-credential]",
		"password", "[redacted-password]",
		"providerSessionId", "[redacted-provider-session]",
		"secret", "[redacted-secret]",
		"token", "[redacted-token]",
	)
	return replacer.Replace(value)
}

var realtimeVoiceDiagnosticAssignmentPattern = regexp.MustCompile(`(?i)\b(api[-_ ]?key|authorization|credential|password|provider[-_ ]?session[-_ ]?id|secret|token)\s*[:=]\s*["']?[^"',\s}\n]+`)
var realtimeVoiceDiagnosticBearerPattern = regexp.MustCompile(`(?i)\b(bearer)\s+[a-z0-9._~+/=-]+`)

func redactRealtimeVoiceDiagnosticValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		redacted := map[string]any{}
		for key, nested := range typed {
			if unsafeRealtimeVoiceDiagnosticKey(key) {
				redacted[key] = "[redacted]"
				continue
			}
			redacted[key] = redactRealtimeVoiceDiagnosticValue(nested)
		}
		return redacted
	case []any:
		redacted := make([]any, 0, len(typed))
		for _, nested := range typed {
			redacted = append(redacted, redactRealtimeVoiceDiagnosticValue(nested))
		}
		return redacted
	case string:
		return redactRealtimeVoiceDiagnosticString(typed)
	default:
		return typed
	}
}

func unsafeRealtimeVoiceDiagnosticKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""), " ", ""))
	for _, token := range []string{"apikey", "authorization", "bearer", "credential", "password", "providersessionid", "secret", "token"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func (a App) ensureRealtimeVoiceDependencies() error {
	if a.authorizer == nil || a.tenants == nil || a.inventories == nil || a.assets == nil || a.search == nil || a.realtimeVoiceProviders == nil || a.realtimeSessions == nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func recoverableRealtimeVoiceToolError(err error) bool {
	return errors.Is(err, ports.ErrInvalidProviderInput) ||
		errors.Is(err, apperrors.ErrInvalidInput) ||
		errors.Is(err, ErrValidation)
}

func (a App) MarkRealtimeVoiceSessionFailed(ctx context.Context, session RealtimeVoiceSession, safeFailureCode string) error {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	return a.markRealtimeVoiceSessionOutcome(ctx, session, ports.RealtimeSessionStateFailed, safeFailureCode)
}

func (a App) MarkRealtimeVoiceSessionCompleted(ctx context.Context, session RealtimeVoiceSession) error {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	return a.markRealtimeVoiceSessionOutcome(ctx, session, ports.RealtimeSessionStateCompleted, "")
}

func (a App) MarkRealtimeVoiceSessionCancelled(ctx context.Context, session RealtimeVoiceSession) error {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	return a.markRealtimeVoiceSessionOutcome(ctx, session, ports.RealtimeSessionStateCancelled, "")
}

func (a App) markRealtimeVoiceSessionOutcome(ctx context.Context, session RealtimeVoiceSession, state ports.RealtimeSessionState, safeFailureCode string) error {
	if a.realtimeSessions == nil || strings.TrimSpace(session.ID) == "" {
		return apperrors.ErrInvalidInput
	}
	return a.realtimeSessions.UpdateRealtimeSessionOutcome(ctx, session.TenantID, session.InventoryID, session.ID, ports.RealtimeSessionOutcome{
		State:           state,
		At:              a.clock.Now(),
		SafeFailureCode: strings.TrimSpace(safeFailureCode),
	})
}

func (a App) newRealtimeVoiceID() string {
	if a.ids == nil {
		return ""
	}
	return a.ids.NewID()
}

func RealtimeVoiceSafeErrorCode(err error) string {
	return realtimeVoiceErrorCode(err)
}
