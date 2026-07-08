package httpserver

import (
	"context"
	"encoding/json"
	"strings"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func readRealtimeAudioMessage(ctx context.Context, connection *websocket.Conn) (realtimeClientMessage, error) {
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
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return realtimeClientMessage{}, err
	}
	messageTypeName := realtimeClientMessageType(strings.TrimSpace(envelope.Type))
	for field := range raw {
		if !realtimeAudioMessageFieldAllowed(messageTypeName, field) {
			return realtimeClientMessage{}, ports.ErrInvalidProviderInput
		}
	}
	if messageTypeName == realtimeClientMessageAudioChunk {
		if err := validateRealtimeAudioChunkFinalMarker(raw); err != nil {
			return realtimeClientMessage{}, err
		}
	}
	var message struct {
		Type         string `json:"type"`
		Seq          int    `json:"seq"`
		SessionID    string `json:"sessionId"`
		ChunkID      string `json:"chunkId"`
		AudioBase64  string `json:"audioBase64"`
		IsFinalChunk bool   `json:"isFinalChunk"`
		Reason       string `json:"reason"`
		AckSeq       int    `json:"ackSeq"`
	}
	if err := json.Unmarshal(payload, &message); err != nil {
		return realtimeClientMessage{}, err
	}
	return realtimeClientMessage{
		Type:         realtimeClientMessageType(strings.TrimSpace(message.Type)),
		Seq:          message.Seq,
		SessionID:    message.SessionID,
		ChunkID:      message.ChunkID,
		AudioBase64:  message.AudioBase64,
		IsFinalChunk: message.IsFinalChunk,
		Reason:       message.Reason,
		AckSeq:       message.AckSeq,
	}, nil
}

func realtimeAudioMessageFieldAllowed(messageType realtimeClientMessageType, field string) bool {
	switch messageType {
	case realtimeClientMessageAudioChunk:
		switch field {
		case "type", "seq", "sessionId", "chunkId", "audioBase64", "isFinalChunk":
			return true
		}
	case realtimeClientMessageAudioEnd:
		switch field {
		case "type", "seq", "sessionId":
			return true
		}
	case realtimeClientMessageSessionCancel:
		switch field {
		case "type", "seq", "sessionId", "reason":
			return true
		}
	case realtimeClientMessageClientAck:
		switch field {
		case "type", "seq", "sessionId", "ackSeq":
			return true
		}
	}
	return false
}

func validateRealtimeAudioChunkFinalMarker(raw map[string]json.RawMessage) error {
	value, ok := raw["isFinalChunk"]
	if !ok {
		return ports.ErrInvalidProviderInput
	}
	var marker *bool
	if err := json.Unmarshal(value, &marker); err != nil {
		return ports.ErrInvalidProviderInput
	}
	if marker == nil {
		return ports.ErrInvalidProviderInput
	}
	return nil
}
