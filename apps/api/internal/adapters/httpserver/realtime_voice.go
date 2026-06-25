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
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeVoicePath = "/v1/realtime/voice"
const maxRealtimeAudioChunkBytes = 512 * 1024
const maxRealtimeVoiceFrameBytes = 710 * 1024

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
			OutputAudio: app.RealtimeVoiceOutputAudio{MimeTypes: start.OutputAudio.MimeTypes},
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

		audioChunks, err := readRealtimeAudio(ctx, connection, session.ID, start.Seq)
		if err != nil {
			_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.failed", Seq: serverSeq, SessionID: session.ID, Code: app.RealtimeVoiceSafeErrorCode(err), Message: "The voice session could not continue."})
			_ = connection.Close(websocket.StatusPolicyViolation, "voice session failed")
			return
		}

		err = application.RunRealtimeVoiceQuery(ctx, app.RealtimeVoiceQueryInput{
			Session:     session,
			AudioChunks: audioChunks,
		}, func(event app.RealtimeVoiceEvent) error {
			message := realtimeServerMessageFromEvent(event, serverSeq)
			serverSeq++
			return writeRealtimeServerMessage(ctx, connection, message)
		})
		if err != nil {
			_ = writeRealtimeServerMessage(ctx, connection, realtimeServerMessage{Type: "session.failed", Seq: serverSeq, SessionID: session.ID, Code: app.RealtimeVoiceSafeErrorCode(err), Message: "The voice session failed safely."})
			_ = connection.Close(websocket.StatusInternalError, "voice session failed")
			return
		}
		_ = connection.Close(websocket.StatusNormalClosure, "voice session completed")
	}
}

type realtimeClientMessage struct {
	Type                  string                 `json:"type"`
	Seq                   int                    `json:"seq"`
	SessionID             string                 `json:"sessionId"`
	TenantID              string                 `json:"tenantId"`
	InventoryID           string                 `json:"inventoryId"`
	Source                string                 `json:"source"`
	RequestedCapabilities []string               `json:"requestedCapabilities"`
	InputAudio            realtimeInputAudio     `json:"inputAudio"`
	OutputAudio           realtimeOutputAudio    `json:"outputAudio"`
	ChunkID               string                 `json:"chunkId"`
	AudioBase64           string                 `json:"audioBase64"`
	IsFinalChunk          bool                   `json:"isFinalChunk"`
	Reason                string                 `json:"reason"`
	Arguments             map[string]interface{} `json:"arguments"`
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
	Type                 string                       `json:"type"`
	Seq                  int                          `json:"seq"`
	SessionID            string                       `json:"sessionId,omitempty"`
	Code                 string                       `json:"code,omitempty"`
	Message              string                       `json:"message,omitempty"`
	Text                 string                       `json:"text,omitempty"`
	Status               string                       `json:"status,omitempty"`
	ToolCallID           string                       `json:"toolCallId,omitempty"`
	ToolLabel            string                       `json:"toolLabel,omitempty"`
	ResponseID           string                       `json:"responseId,omitempty"`
	Response             *realtimeStructuredResponse  `json:"response,omitempty"`
	Format               *realtimeAudioFormatResponse `json:"format,omitempty"`
	ChunkID              string                       `json:"chunkId,omitempty"`
	AudioBase64          string                       `json:"audioBase64,omitempty"`
	IsFinalChunk         bool                         `json:"isFinalChunk,omitempty"`
	AcceptedInputAudio   realtimeInputAudio           `json:"acceptedInputAudio,omitempty"`
	AcceptedOutputAudio  realtimeOutputAudio          `json:"acceptedOutputAudio,omitempty"`
	AcceptedCapabilities []string                     `json:"acceptedCapabilities,omitempty"`
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

func writeRealtimeServerMessage(ctx context.Context, connection *websocket.Conn, message realtimeServerMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return connection.Write(ctx, websocket.MessageText, payload)
}

func readRealtimeAudio(ctx context.Context, connection *websocket.Conn, sessionID string, lastClientSeq int) ([][]byte, error) {
	chunks := [][]byte{}
	seenChunks := map[string]struct{}{}
	for {
		message, err := readRealtimeClientMessage(ctx, connection)
		if err != nil {
			return nil, err
		}
		if message.Seq <= lastClientSeq {
			return nil, ports.ErrInvalidProviderInput
		}
		lastClientSeq = message.Seq
		if message.SessionID != sessionID {
			return nil, ports.ErrForbidden
		}
		switch message.Type {
		case "audio.chunk":
			chunkID := strings.TrimSpace(message.ChunkID)
			if chunkID == "" {
				return nil, ports.ErrInvalidProviderInput
			}
			if _, exists := seenChunks[chunkID]; exists {
				return nil, ports.ErrInvalidProviderInput
			}
			seenChunks[chunkID] = struct{}{}
			chunk, err := base64.StdEncoding.DecodeString(message.AudioBase64)
			if err != nil || len(chunk) == 0 || len(chunk) > maxRealtimeAudioChunkBytes {
				return nil, ports.ErrInvalidProviderInput
			}
			chunks = append(chunks, chunk)
		case "audio.end":
			if len(chunks) == 0 {
				return nil, ports.ErrInvalidProviderInput
			}
			return chunks, nil
		case "session.cancel":
			return nil, errors.New("voice session cancelled")
		default:
			return nil, ports.ErrInvalidProviderInput
		}
	}
}

func realtimeServerMessageFromEvent(event app.RealtimeVoiceEvent, seq int) realtimeServerMessage {
	message := realtimeServerMessage{
		Type:       event.Type,
		Seq:        seq,
		SessionID:  event.SessionID,
		Status:     event.Status,
		Code:       event.Code,
		Message:    event.Message,
		Text:       event.Text,
		ToolCallID: event.ToolCallID,
		ToolLabel:  event.ToolLabel,
		ChunkID:    event.ChunkID,
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
	case app.RealtimeVoiceEventTextToSpeechAudioStarted:
		message.Format = &realtimeAudioFormatResponse{MimeType: event.AudioMime}
	case app.RealtimeVoiceEventTextToSpeechAudioChunk:
		message.AudioBase64 = base64.StdEncoding.EncodeToString(event.Audio)
		message.IsFinalChunk = event.FinalChunk
	}
	return message
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
