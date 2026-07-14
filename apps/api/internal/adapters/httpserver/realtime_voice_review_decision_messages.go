package httpserver

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

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
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return realtimeClientMessage{}, err
	}
	messageTypeName := realtimeClientMessageType(strings.TrimSpace(envelope.Type))
	for field := range raw {
		if !realtimeActionPlanDecisionFieldAllowed(messageTypeName, field) {
			return realtimeClientMessage{}, ports.ErrInvalidProviderInput
		}
	}
	if err := validateRealtimePhotoAttachmentRequestFields(raw); err != nil {
		return realtimeClientMessage{}, err
	}
	if err := validateRealtimeCommandEditFields(raw); err != nil {
		return realtimeClientMessage{}, err
	}
	var message struct {
		Type             string                           `json:"type"`
		Seq              int                              `json:"seq"`
		SessionID        string                           `json:"sessionId"`
		PlanID           string                           `json:"planId"`
		AckSeq           int                              `json:"ackSeq"`
		PhotoAttachments []realtimePhotoAttachmentRequest `json:"photoAttachments,omitempty"`
		CommandEdits     []realtimeActionPlanCommandEdit  `json:"commandEdits,omitempty"`
	}
	if err := json.Unmarshal(payload, &message); err != nil {
		return realtimeClientMessage{}, err
	}
	_, photoAttachmentsSet := raw["photoAttachments"]
	_, commandEditsSet := raw["commandEdits"]
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
		Type:                realtimeClientMessageType(strings.TrimSpace(message.Type)),
		Seq:                 message.Seq,
		SessionID:           message.SessionID,
		PlanID:              message.PlanID,
		AckSeq:              message.AckSeq,
		PhotoAttachments:    message.PhotoAttachments,
		PhotoAttachmentsSet: photoAttachmentsSet,
		CommandEdits:        message.CommandEdits,
		CommandEditsSet:     commandEditsSet,
	}, nil
}

func validateRealtimeCommandEditFields(raw map[string]json.RawMessage) error {
	value, ok := raw["commandEdits"]
	if !ok {
		return nil
	}
	var edits []map[string]json.RawMessage
	if err := json.Unmarshal(value, &edits); err != nil || len(edits) > 10 {
		return ports.ErrInvalidProviderInput
	}
	seen := map[string]struct{}{}
	for _, rawEdit := range edits {
		for field := range rawEdit {
			if field != "commandId" && field != "title" && field != "parent" {
				return ports.ErrInvalidProviderInput
			}
		}
		var edit realtimeActionPlanCommandEdit
		encoded, _ := json.Marshal(rawEdit)
		if json.Unmarshal(encoded, &edit) != nil || strings.TrimSpace(edit.CommandID) == "" || (edit.Title == nil && edit.Parent == nil) {
			return ports.ErrInvalidProviderInput
		}
		if _, duplicate := seen[strings.TrimSpace(edit.CommandID)]; duplicate {
			return ports.ErrInvalidProviderInput
		}
		seen[strings.TrimSpace(edit.CommandID)] = struct{}{}
		if edit.Title != nil && (strings.TrimSpace(*edit.Title) == "" || len([]rune(strings.TrimSpace(*edit.Title))) > 200) {
			return ports.ErrInvalidProviderInput
		}
		if edit.Parent != nil {
			var parentFields map[string]json.RawMessage
			if json.Unmarshal(rawEdit["parent"], &parentFields) != nil {
				return ports.ErrInvalidProviderInput
			}
			for field := range parentFields {
				if field != "kind" && field != "id" {
					return ports.ErrInvalidProviderInput
				}
			}
			kind, id := strings.TrimSpace(edit.Parent.Kind), strings.TrimSpace(edit.Parent.ID)
			if (kind == "root" && id != "") || ((kind == "asset" || kind == "command") && id == "") || (kind != "root" && kind != "asset" && kind != "command") {
				return ports.ErrInvalidProviderInput
			}
		}
	}
	return nil
}

func validateRealtimePhotoAttachmentRequestFields(raw map[string]json.RawMessage) error {
	value, ok := raw["photoAttachments"]
	if !ok {
		return nil
	}
	var photos []map[string]json.RawMessage
	if err := json.Unmarshal(value, &photos); err != nil {
		return ports.ErrInvalidProviderInput
	}
	for _, photo := range photos {
		for field := range photo {
			if !realtimePhotoAttachmentRequestFieldAllowed(field) {
				return ports.ErrInvalidProviderInput
			}
		}
	}
	return nil
}

func realtimePhotoAttachmentRequestFieldAllowed(field string) bool {
	switch field {
	case "commandId", "photoIndex", "fileName", "contentType", "sizeBytes":
		return true
	default:
		return false
	}
}

func realtimeActionPlanDecisionFieldAllowed(messageType realtimeClientMessageType, field string) bool {
	switch messageType {
	case realtimeClientMessageActionPlanApprove, realtimeClientMessageActionPlanCancel:
		switch field {
		case "type", "seq", "sessionId", "planId", "photoAttachments", "commandEdits":
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
