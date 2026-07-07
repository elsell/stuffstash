package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeVoicePath = "/v1/realtime/voice"
const maxRealtimeAudioChunkBytes = 512 * 1024
const maxRealtimeVoiceFrameBytes = 710 * 1024
const maxRealtimeVoiceTurnsPerSession = 3

var errRealtimeVoiceCancelled = errors.New("voice session cancelled")

func handleRealtimeVoice(application app.App, sessionTimeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		principal, err := application.Authenticate(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer connection.Close(websocket.StatusInternalError, "voice session ended")
		connection.SetReadLimit(maxRealtimeVoiceFrameBytes)

		ctx, cancelSession := context.WithTimeout(r.Context(), sessionTimeout)
		defer cancelSession()
		start, err := readRealtimeClientMessage(ctx, connection)
		if err != nil {
			_ = connection.Close(websocket.StatusPolicyViolation, "invalid start message")
			return
		}
		if start.Type != "session.start" {
			_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.failed", Seq: 1, Code: "invalid_request", Message: "The voice session could not be started."})
			return
		}
		session, err := application.StartRealtimeVoiceSession(ctx, app.RealtimeVoiceSessionInput{
			Principal:   principal,
			TenantID:    tenant.ID(start.TenantID),
			InventoryID: inventory.InventoryID(start.InventoryID),
			Source:      start.Source,
			InputAudio: ports.RealtimeAudioFormat{
				MimeType:   start.InputAudio.MimeType,
				SampleRate: start.InputAudio.SampleRate,
				Channels:   start.InputAudio.Channels,
			},
			OutputAudio:          app.RealtimeVoiceOutputAudio{MimeTypes: start.OutputAudio.MimeTypes},
			DeveloperDiagnostics: start.DeveloperDiagnostics,
		})
		serverSeq := 1
		if err != nil {
			_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.failed", Seq: serverSeq, Code: app.RealtimeVoiceSafeErrorCode(err), Message: "The voice session could not be started."})
			_ = connection.Close(websocket.StatusPolicyViolation, "voice session rejected")
			return
		}
		if err := writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{
			Type:                 "session.started",
			Seq:                  serverSeq,
			SessionID:            session.ID,
			AcceptedInputAudio:   start.InputAudio,
			AcceptedOutputAudio:  start.OutputAudio,
			AcceptedCapabilities: start.RequestedCapabilities,
		}); err != nil {
			return
		}
		serverSeq++

		lastClientSeq := start.Seq
		conversationTurns := []ports.AgentConversationTurn{}
		seenAudioChunkIDs := map[string]struct{}{}
		for turn := 0; turn < maxRealtimeVoiceTurnsPerSession; turn++ {
			audioChunks, nextClientSeq, err := readRealtimeAudio(ctx, connection, session.ID, lastClientSeq, seenAudioChunkIDs)
			lastClientSeq = nextClientSeq
			if err != nil {
				if errors.Is(err, errRealtimeVoiceCancelled) {
					_ = application.MarkRealtimeVoiceSessionCancelled(ctx, session)
					_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.cancelled", Seq: serverSeq, SessionID: session.ID})
					_ = connection.Close(websocket.StatusNormalClosure, "voice session cancelled")
					return
				} else {
					_ = application.MarkRealtimeVoiceSessionFailed(ctx, session, app.RealtimeVoiceSafeErrorCode(err))
				}
				_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.failed", Seq: serverSeq, SessionID: session.ID, Code: app.RealtimeVoiceSafeErrorCode(err), Message: "The voice session could not continue."})
				_ = connection.Close(websocket.StatusPolicyViolation, "voice session failed")
				return
			}

			reviewPlanID := ""
			completedResponseKind := ""
			turnTranscript := ""
			var completedResponse *ports.StructuredAgentResponse
			err = application.RunRealtimeVoiceQuery(ctx, app.RealtimeVoiceQueryInput{
				Session:                    session,
				AudioChunks:                audioChunks,
				ContinueAfterClarification: true,
				ConversationTurns:          conversationTurns,
			}, func(event app.RealtimeVoiceEvent) error {
				if event.Type == app.RealtimeVoiceEventActionPlanProposed && event.ActionPlan != nil {
					reviewPlanID = strings.TrimSpace(event.ActionPlan.PlanID)
				}
				if event.Type == app.RealtimeVoiceEventTranscriptFinal {
					turnTranscript = strings.TrimSpace(event.Text)
				}
				if event.Type == app.RealtimeVoiceEventAssistantResponseCompleted && event.Response != nil {
					completedResponseKind = string(event.Response.Kind)
					responseCopy := *event.Response
					completedResponse = &responseCopy
				}
				message := realtimeServerMessageFromEvent(event, serverSeq)
				serverSeq++
				return writeRealtimeServerMessage(ctx, connection, message)
			})
			if err != nil {
				if errors.Is(err, context.Canceled) {
					_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.cancelled", Seq: serverSeq, SessionID: session.ID})
					_ = connection.Close(websocket.StatusNormalClosure, "voice session cancelled")
					return
				}
				_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.failed", Seq: serverSeq, SessionID: session.ID, Code: app.RealtimeVoiceSafeErrorCode(err), Message: "The voice session failed safely."})
				_ = connection.Close(websocket.StatusInternalError, "voice session failed")
				return
			}
			if reviewPlanID != "" {
				reviewOutcome, err := handleRealtimeActionPlanDecision(ctx, connection, application, session, reviewPlanID, &lastClientSeq, &serverSeq)
				if err != nil {
					_ = application.MarkRealtimeVoiceSessionFailed(ctx, session, app.RealtimeVoiceSafeErrorCode(err))
					_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.failed", Seq: serverSeq, SessionID: session.ID, Code: app.RealtimeVoiceSafeErrorCode(err), Message: "The action plan decision could not be applied safely."})
					_ = connection.Close(websocket.StatusPolicyViolation, "voice review failed")
					return
				}
				switch reviewOutcome {
				case ports.RealtimeSessionStateCancelled:
					_ = application.MarkRealtimeVoiceSessionCancelled(ctx, session)
				default:
					_ = application.MarkRealtimeVoiceSessionCompleted(ctx, session)
				}
				_ = connection.Close(websocket.StatusNormalClosure, "voice session completed")
				return
			}
			conversationTurns = appendRealtimeVoiceConversationTurns(conversationTurns, turnTranscript, completedResponse)
			if completedResponseKind != string(ports.StructuredAgentResponseKindClarification) {
				_ = connection.Close(websocket.StatusNormalClosure, "voice session completed")
				return
			}
		}
		_ = application.MarkRealtimeVoiceSessionFailed(ctx, session, "clarification_turn_limit")
		_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.failed", Seq: serverSeq, SessionID: session.ID, Code: "clarification_turn_limit", Message: "The voice session needs a fresh start."})
		_ = connection.Close(websocket.StatusNormalClosure, "voice session completed")
	}
}

func appendRealtimeVoiceConversationTurns(turns []ports.AgentConversationTurn, transcript string, response *ports.StructuredAgentResponse) []ports.AgentConversationTurn {
	transcript = strings.TrimSpace(transcript)
	if transcript != "" {
		turns = append(turns, ports.AgentConversationTurn{Role: ports.AgentConversationRoleUser, Text: transcript})
	}
	if response != nil {
		text := strings.TrimSpace(response.DisplayResponse)
		if text == "" {
			text = strings.TrimSpace(response.SpokenResponse)
		}
		if text != "" {
			turns = append(turns, ports.AgentConversationTurn{
				Role: ports.AgentConversationRoleAssistant,
				Kind: string(response.Kind),
				Text: text,
			})
		}
	}
	const maxTurns = 6
	if len(turns) > maxTurns {
		return append([]ports.AgentConversationTurn{}, turns[len(turns)-maxTurns:]...)
	}
	return append([]ports.AgentConversationTurn{}, turns...)
}

type realtimeClientMessage struct {
	Type                  string                           `json:"type"`
	Seq                   int                              `json:"seq"`
	SessionID             string                           `json:"sessionId"`
	TenantID              string                           `json:"tenantId"`
	InventoryID           string                           `json:"inventoryId"`
	Source                string                           `json:"source"`
	RequestedCapabilities []string                         `json:"requestedCapabilities"`
	InputAudio            realtimeInputAudio               `json:"inputAudio"`
	OutputAudio           realtimeOutputAudio              `json:"outputAudio"`
	DeveloperDiagnostics  bool                             `json:"developerDiagnostics,omitempty"`
	ChunkID               string                           `json:"chunkId"`
	PlanID                string                           `json:"planId"`
	AudioBase64           string                           `json:"audioBase64"`
	IsFinalChunk          bool                             `json:"isFinalChunk"`
	Reason                string                           `json:"reason"`
	Arguments             map[string]interface{}           `json:"arguments"`
	PhotoAttachments      []realtimePhotoAttachmentRequest `json:"photoAttachments,omitempty"`
}

type realtimeInputAudio struct {
	MimeType   string `json:"mimeType"`
	SampleRate int    `json:"sampleRate"`
	Channels   int    `json:"channels"`
}

type realtimeOutputAudio struct {
	MimeTypes []string `json:"mimeTypes"`
}

type realtimeServerMessage struct {
	Type                    string                            `json:"type"`
	Seq                     int                               `json:"seq"`
	SessionID               string                            `json:"sessionId,omitempty"`
	Code                    string                            `json:"code,omitempty"`
	Message                 string                            `json:"message,omitempty"`
	Text                    string                            `json:"text,omitempty"`
	Detail                  string                            `json:"detail,omitempty"`
	Status                  string                            `json:"status,omitempty"`
	ToolCallID              string                            `json:"toolCallId,omitempty"`
	ToolLabel               string                            `json:"toolLabel,omitempty"`
	ActionPlan              *realtimeActionPlanProposal       `json:"actionPlan,omitempty"`
	PlanID                  string                            `json:"planId,omitempty"`
	CommandResults          []realtimeActionPlanCommandResult `json:"commandResults,omitempty"`
	AttachmentUploadIntents []realtimeAttachmentUploadIntent  `json:"attachmentUploadIntents,omitempty"`
	ResponseID              string                            `json:"responseId,omitempty"`
	Response                *realtimeStructuredResponse       `json:"response,omitempty"`
	Format                  *realtimeAudioFormatResponse      `json:"format,omitempty"`
	ChunkID                 string                            `json:"chunkId,omitempty"`
	AudioBase64             string                            `json:"audioBase64,omitempty"`
	IsFinalChunk            bool                              `json:"isFinalChunk,omitempty"`
	AcceptedInputAudio      realtimeInputAudio                `json:"acceptedInputAudio,omitempty"`
	AcceptedOutputAudio     realtimeOutputAudio               `json:"acceptedOutputAudio,omitempty"`
	AcceptedCapabilities    []string                          `json:"acceptedCapabilities,omitempty"`
}

type realtimeStructuredResponse struct {
	ResponseID      string   `json:"responseId"`
	SessionID       string   `json:"sessionId"`
	TenantID        string   `json:"tenantId"`
	InventoryID     string   `json:"inventoryId"`
	Source          string   `json:"source"`
	Kind            string   `json:"kind"`
	SpokenResponse  string   `json:"spokenResponse"`
	DisplayResponse string   `json:"displayResponse"`
	Artifacts       []any    `json:"artifacts"`
	ToolCallIDs     []string `json:"toolCallIds"`
	AuditMetadata   struct{} `json:"auditMetadata"`
}

type realtimeActionPlanProposal struct {
	PlanID              string                      `json:"planId"`
	ConfirmationSummary string                      `json:"confirmationSummary"`
	Commands            []realtimeActionPlanCommand `json:"commands"`
	Risks               []string                    `json:"risks"`
}

type realtimeActionPlanCommand struct {
	ID              string `json:"id,omitempty"`
	Kind            string `json:"kind"`
	Summary         string `json:"summary"`
	Operation       string `json:"operation,omitempty"`
	Title           string `json:"title,omitempty"`
	AssetKind       string `json:"assetKind,omitempty"`
	ParentAssetID   string `json:"parentAssetId,omitempty"`
	ParentTitle     string `json:"parentTitle,omitempty"`
	ParentKind      string `json:"parentKind,omitempty"`
	ParentCommandID string `json:"parentCommandId,omitempty"`
}

type realtimeActionPlanCommandResult struct {
	CommandID string `json:"commandId"`
	AssetID   string `json:"assetId"`
	Operation string `json:"operation"`
	AssetKind string `json:"assetKind"`
}

type realtimePhotoAttachmentRequest struct {
	CommandID   string `json:"commandId"`
	PhotoIndex  int    `json:"photoIndex"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

type realtimeAttachmentUploadIntent struct {
	CommandID    string                     `json:"commandId"`
	PhotoIndex   int                        `json:"photoIndex"`
	AssetID      string                     `json:"assetId"`
	FileName     string                     `json:"fileName"`
	ContentType  string                     `json:"contentType"`
	SizeBytes    int64                      `json:"sizeBytes"`
	DirectUpload realtimeDirectUploadIntent `json:"directUpload"`
}

type realtimeDirectUploadIntent struct {
	UploadID     string            `json:"uploadId"`
	AttachmentID string            `json:"attachmentId"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	FormFields   map[string]string `json:"formFields"`
	ExpiresAt    string            `json:"expiresAt"`
}

type realtimeAudioFormatResponse struct {
	MimeType string `json:"mimeType"`
}

func readRealtimeClientMessage(ctx context.Context, connection *websocket.Conn) (realtimeClientMessage, error) {
	messageType, payload, err := connection.Read(ctx)
	if err != nil {
		return realtimeClientMessage{}, err
	}
	if messageType != websocket.MessageText {
		return realtimeClientMessage{}, ports.ErrInvalidProviderInput
	}
	var message realtimeClientMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return realtimeClientMessage{}, err
	}
	message.Type = strings.TrimSpace(message.Type)
	return message, nil
}

func readRealtimeActionPlanDecisionMessage(ctx context.Context, connection *websocket.Conn) (realtimeClientMessage, error) {
	messageType, payload, err := connection.Read(ctx)
	if err != nil {
		return realtimeClientMessage{}, err
	}
	if messageType != websocket.MessageText {
		return realtimeClientMessage{}, ports.ErrInvalidProviderInput
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return realtimeClientMessage{}, err
	}
	for field := range raw {
		switch field {
		case "type", "seq", "sessionId", "planId", "photoAttachments":
		default:
			return realtimeClientMessage{}, ports.ErrInvalidProviderInput
		}
	}
	var message struct {
		Type             string                           `json:"type"`
		Seq              int                              `json:"seq"`
		SessionID        string                           `json:"sessionId"`
		PlanID           string                           `json:"planId"`
		PhotoAttachments []realtimePhotoAttachmentRequest `json:"photoAttachments,omitempty"`
	}
	if err := json.Unmarshal(payload, &message); err != nil {
		return realtimeClientMessage{}, err
	}
	if len(message.PhotoAttachments) > 10 {
		return realtimeClientMessage{}, ports.ErrInvalidProviderInput
	}
	seenPhotos := map[string]struct{}{}
	for _, photo := range message.PhotoAttachments {
		commandID := strings.TrimSpace(photo.CommandID)
		if commandID == "" || photo.PhotoIndex < 0 || photo.PhotoIndex > 9 || strings.TrimSpace(photo.FileName) == "" || strings.TrimSpace(photo.ContentType) == "" || photo.SizeBytes <= 0 {
			return realtimeClientMessage{}, ports.ErrInvalidProviderInput
		}
		key := commandID + ":" + strconv.Itoa(photo.PhotoIndex)
		if _, exists := seenPhotos[key]; exists {
			return realtimeClientMessage{}, ports.ErrInvalidProviderInput
		}
		seenPhotos[key] = struct{}{}
	}
	return realtimeClientMessage{
		Type:             strings.TrimSpace(message.Type),
		Seq:              message.Seq,
		SessionID:        message.SessionID,
		PlanID:           message.PlanID,
		PhotoAttachments: message.PhotoAttachments,
	}, nil
}

func writeRealtimeServerMessage(ctx context.Context, connection *websocket.Conn, message realtimeServerMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return connection.Write(ctx, websocket.MessageText, payload)
}

func readRealtimeAudio(ctx context.Context, connection *websocket.Conn, sessionID string, lastClientSeq int, seenSessionChunkIDs map[string]struct{}) ([][]byte, int, error) {
	chunks := [][]byte{}
	seenChunks := map[string]struct{}{}
	for {
		message, err := readRealtimeAudioMessage(ctx, connection)
		if err != nil {
			return nil, lastClientSeq, err
		}
		if message.Seq <= lastClientSeq {
			return nil, lastClientSeq, ports.ErrInvalidProviderInput
		}
		lastClientSeq = message.Seq
		if message.SessionID != sessionID {
			return nil, lastClientSeq, ports.ErrForbidden
		}
		switch message.Type {
		case "audio.chunk":
			chunkID := strings.TrimSpace(message.ChunkID)
			if chunkID == "" {
				return nil, lastClientSeq, ports.ErrInvalidProviderInput
			}
			if _, exists := seenChunks[chunkID]; exists {
				return nil, lastClientSeq, ports.ErrInvalidProviderInput
			}
			if _, exists := seenSessionChunkIDs[chunkID]; exists {
				return nil, lastClientSeq, ports.ErrInvalidProviderInput
			}
			seenChunks[chunkID] = struct{}{}
			seenSessionChunkIDs[chunkID] = struct{}{}
			chunk, err := base64.StdEncoding.DecodeString(message.AudioBase64)
			if err != nil || len(chunk) == 0 || len(chunk) > maxRealtimeAudioChunkBytes {
				return nil, lastClientSeq, ports.ErrInvalidProviderInput
			}
			chunks = append(chunks, chunk)
		case "audio.end":
			if len(chunks) == 0 {
				return nil, lastClientSeq, ports.ErrInvalidProviderInput
			}
			return chunks, lastClientSeq, nil
		case "session.cancel":
			return nil, lastClientSeq, errRealtimeVoiceCancelled
		default:
			return nil, lastClientSeq, ports.ErrInvalidProviderInput
		}
	}
}

func handleRealtimeActionPlanDecision(ctx context.Context, connection *websocket.Conn, application app.App, session app.RealtimeVoiceSession, expectedPlanID string, lastClientSeq *int, serverSeq *int) (ports.RealtimeSessionState, error) {
	message, err := readRealtimeActionPlanDecisionMessage(ctx, connection)
	if err != nil {
		return "", err
	}
	if message.Seq <= *lastClientSeq {
		return "", ports.ErrInvalidProviderInput
	}
	*lastClientSeq = message.Seq
	if message.SessionID != session.ID {
		return "", ports.ErrForbidden
	}
	planID := strings.TrimSpace(message.PlanID)
	if planID == "" || planID != strings.TrimSpace(expectedPlanID) {
		return "", ports.ErrForbidden
	}

	var eventType string
	var status string
	switch message.Type {
	case "action.plan.approve":
		if err := application.ValidateActionPlanPhotoAttachmentMetadata(ctx, app.ActionPlanPhotoAttachmentMetadataInput{
			Decision: app.ActionPlanDecisionInput{
				Principal:   session.Principal,
				TenantID:    session.TenantID,
				InventoryID: session.InventoryID,
				PlanID:      planID,
			},
			Photos: realtimePhotoAttachmentMetadataFromRequests(message.PhotoAttachments),
		}); err != nil {
			return "", err
		}
		record, err := application.ApproveActionPlan(ctx, app.ActionPlanDecisionInput{
			Principal:   session.Principal,
			TenantID:    session.TenantID,
			InventoryID: session.InventoryID,
			PlanID:      planID,
		})
		if err != nil {
			return "", err
		}
		eventType = app.RealtimeVoiceEventActionPlanApproved
		status = string(record.State)
		if err := writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: eventType, Seq: *serverSeq, SessionID: session.ID, PlanID: planID, Status: status}); err != nil {
			return "", err
		}
		*serverSeq = *serverSeq + 1

		executed, err := application.ExecuteActionPlanDetailed(ctx, app.ActionPlanDecisionInput{
			Principal:   session.Principal,
			TenantID:    session.TenantID,
			InventoryID: session.InventoryID,
			PlanID:      planID,
		})
		outcomeType := app.RealtimeVoiceEventActionPlanExecuted
		outcomeMessage := "The approved change was applied."
		if err != nil {
			if executed.Record.State != actionplan.StateFailed {
				return "", err
			}
			outcomeType = app.RealtimeVoiceEventActionPlanFailed
			outcomeMessage = "The approved change could not be applied safely."
		}
		uploadIntents, err := realtimeAttachmentUploadIntentsFromDecision(ctx, application, session, message.PhotoAttachments, executed.CommandResults)
		if err != nil {
			uploadIntents = nil
			if outcomeType == app.RealtimeVoiceEventActionPlanExecuted {
				outcomeMessage = "The approved change was applied, but photos could not be prepared for upload."
			}
		}
		if err := writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{
			Type:                    outcomeType,
			Seq:                     *serverSeq,
			SessionID:               session.ID,
			PlanID:                  planID,
			Status:                  string(executed.Record.State),
			Message:                 outcomeMessage,
			CommandResults:          realtimeActionPlanCommandResultsFromApp(executed.CommandResults),
			AttachmentUploadIntents: uploadIntents,
		}); err != nil {
			return "", err
		}
		*serverSeq = *serverSeq + 1
		return ports.RealtimeSessionStateCompleted, nil
	case "action.plan.cancel":
		record, err := application.CancelActionPlan(ctx, app.ActionPlanDecisionInput{
			Principal:   session.Principal,
			TenantID:    session.TenantID,
			InventoryID: session.InventoryID,
			PlanID:      planID,
		})
		if err != nil {
			return "", err
		}
		eventType = app.RealtimeVoiceEventActionPlanCancelled
		status = string(record.State)
	default:
		return "", ports.ErrInvalidProviderInput
	}

	if err := writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: eventType, Seq: *serverSeq, SessionID: session.ID, PlanID: planID, Status: status}); err != nil {
		return "", err
	}
	*serverSeq = *serverSeq + 1
	return ports.RealtimeSessionStateCancelled, nil
}

func realtimePhotoAttachmentMetadataFromRequests(photos []realtimePhotoAttachmentRequest) []app.ActionPlanPhotoAttachmentMetadata {
	if len(photos) == 0 {
		return nil
	}
	metadata := make([]app.ActionPlanPhotoAttachmentMetadata, 0, len(photos))
	for _, photo := range photos {
		metadata = append(metadata, app.ActionPlanPhotoAttachmentMetadata{
			CommandID:   photo.CommandID,
			FileName:    photo.FileName,
			ContentType: photo.ContentType,
			SizeBytes:   photo.SizeBytes,
		})
	}
	return metadata
}

func realtimeAttachmentUploadIntentsFromDecision(ctx context.Context, application app.App, session app.RealtimeVoiceSession, photos []realtimePhotoAttachmentRequest, results []app.ActionPlanCommandExecutionResult) ([]realtimeAttachmentUploadIntent, error) {
	if len(photos) == 0 || len(results) == 0 {
		return nil, nil
	}
	resultsByCommandID := map[string]app.ActionPlanCommandExecutionResult{}
	for _, result := range results {
		if result.CommandID != "" && result.AssetID != "" && photoAttachableCommandResult(result) {
			resultsByCommandID[result.CommandID] = result
		}
	}
	intents := make([]realtimeAttachmentUploadIntent, 0, len(photos))
	for _, photo := range photos {
		commandID := strings.TrimSpace(photo.CommandID)
		result, ok := resultsByCommandID[commandID]
		if !ok {
			continue
		}
		upload, err := application.InitiateAttachmentDirectUpload(ctx, app.InitiateAttachmentDirectUploadInput{
			Principal:   session.Principal,
			Source:      audit.SourceConversation,
			RequestID:   session.ID + ":" + commandID,
			TenantID:    session.TenantID,
			InventoryID: session.InventoryID,
			AssetID:     asset.ID(result.AssetID),
			FileName:    photo.FileName,
			ContentType: photo.ContentType,
			SizeBytes:   photo.SizeBytes,
		})
		if err != nil {
			return nil, err
		}
		intents = append(intents, realtimeAttachmentUploadIntent{
			CommandID:   commandID,
			PhotoIndex:  photo.PhotoIndex,
			AssetID:     result.AssetID,
			FileName:    strings.TrimSpace(photo.FileName),
			ContentType: strings.TrimSpace(photo.ContentType),
			SizeBytes:   photo.SizeBytes,
			DirectUpload: realtimeDirectUploadIntent{
				UploadID:     upload.UploadID,
				AttachmentID: upload.AttachmentID.String(),
				Method:       upload.Method,
				URL:          upload.URL,
				Headers:      copyRealtimeStringMap(upload.Headers),
				FormFields:   copyRealtimeStringMap(upload.FormFields),
				ExpiresAt:    upload.ExpiresAt.UTC().Format(time.RFC3339),
			},
		})
	}
	if len(intents) != len(photos) {
		return nil, ports.ErrInvalidProviderInput
	}
	return intents, nil
}

func photoAttachableCommandResult(result app.ActionPlanCommandExecutionResult) bool {
	return (result.Operation == "create" || result.Operation == "move") &&
		(result.AssetKind == "item" || result.AssetKind == "container" || result.AssetKind == "location")
}

func copyRealtimeStringMap(values map[string]string) map[string]string {
	copied := map[string]string{}
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func realtimeServerMessageFromEvent(event app.RealtimeVoiceEvent, seq int) realtimeServerMessage {
	message := realtimeServerMessage{
		Type:           event.Type,
		Seq:            seq,
		SessionID:      event.SessionID,
		Status:         event.Status,
		Code:           event.Code,
		Message:        event.Message,
		Text:           event.Text,
		Detail:         event.Detail,
		ToolCallID:     event.ToolCallID,
		ToolLabel:      event.ToolLabel,
		PlanID:         event.PlanID,
		CommandResults: realtimeVoiceEventCommandResultsFromApp(event.CommandResults),
		ChunkID:        event.ChunkID,
	}
	switch event.Type {
	case app.RealtimeVoiceEventAssistantResponseStarted:
		if event.Response != nil {
			message.ResponseID = event.Response.ResponseID
		}
	case app.RealtimeVoiceEventAssistantResponseCompleted:
		if event.Response != nil {
			message.Response = realtimeStructuredResponseFromPort(*event.Response)
		}
	case app.RealtimeVoiceEventActionPlanProposed:
		if event.ActionPlan != nil {
			message.ActionPlan = realtimeActionPlanFromApp(*event.ActionPlan)
		}
	case app.RealtimeVoiceEventTextToSpeechAudioStarted:
		message.Format = &realtimeAudioFormatResponse{MimeType: event.AudioMime}
	case app.RealtimeVoiceEventTextToSpeechAudioChunk:
		message.AudioBase64 = base64.StdEncoding.EncodeToString(event.Audio)
		message.IsFinalChunk = event.FinalChunk
	}
	return message
}

func realtimeActionPlanCommandResultsFromApp(results []app.ActionPlanCommandExecutionResult) []realtimeActionPlanCommandResult {
	if len(results) == 0 {
		return nil
	}
	mapped := make([]realtimeActionPlanCommandResult, 0, len(results))
	for _, result := range results {
		mapped = append(mapped, realtimeActionPlanCommandResult{
			CommandID: result.CommandID,
			AssetID:   result.AssetID,
			Operation: result.Operation,
			AssetKind: result.AssetKind,
		})
	}
	return mapped
}

func realtimeVoiceEventCommandResultsFromApp(results []app.RealtimeVoiceActionPlanCommandResult) []realtimeActionPlanCommandResult {
	if len(results) == 0 {
		return nil
	}
	mapped := make([]realtimeActionPlanCommandResult, 0, len(results))
	for _, result := range results {
		mapped = append(mapped, realtimeActionPlanCommandResult{
			CommandID: result.CommandID,
			AssetID:   result.AssetID,
			Operation: result.Operation,
			AssetKind: result.AssetKind,
		})
	}
	return mapped
}

func realtimeActionPlanFromApp(proposal app.RealtimeVoiceActionPlanProposal) *realtimeActionPlanProposal {
	commands := make([]realtimeActionPlanCommand, 0, len(proposal.Commands))
	for _, command := range proposal.Commands {
		commands = append(commands, realtimeActionPlanCommand{
			ID:              command.ID,
			Kind:            command.Kind,
			Summary:         command.Summary,
			Operation:       command.Operation,
			Title:           command.Title,
			AssetKind:       command.AssetKind,
			ParentAssetID:   command.ParentAssetID,
			ParentTitle:     command.ParentTitle,
			ParentKind:      command.ParentKind,
			ParentCommandID: command.ParentCommandID,
		})
	}
	return &realtimeActionPlanProposal{
		PlanID:              proposal.PlanID,
		ConfirmationSummary: proposal.ConfirmationSummary,
		Commands:            commands,
		Risks:               append([]string{}, proposal.Risks...),
	}
}

func realtimeStructuredResponseFromPort(response ports.StructuredAgentResponse) *realtimeStructuredResponse {
	return &realtimeStructuredResponse{
		ResponseID:      response.ResponseID,
		SessionID:       response.SessionID,
		TenantID:        response.TenantID.String(),
		InventoryID:     response.InventoryID.String(),
		Source:          response.Source,
		Kind:            string(response.Kind),
		SpokenResponse:  response.SpokenResponse,
		DisplayResponse: response.DisplayResponse,
		Artifacts:       []any{},
		ToolCallIDs:     response.ToolCallIDs,
	}
}

func realtimeVoiceFailureStatus(code string) int {
	switch code {
	case "unauthenticated":
		return http.StatusUnauthorized
	case "forbidden":
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}

func realtimeVoiceSequenceString(seq int) string {
	return strconv.Itoa(seq)
}
