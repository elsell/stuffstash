package httpserver

import (
	"context"
	"strings"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func handleRealtimeActionPlanDecision(ctx context.Context, connection *websocket.Conn, application app.App, session app.RealtimeVoiceSession, expectedPlanID string, lastClientSeq *int, serverSeq *int) (ports.RealtimeSessionState, error) {
	var message realtimeClientMessage
	for {
		next, err := readRealtimeActionPlanDecisionMessage(ctx, connection)
		if err != nil {
			return "", err
		}
		if next.Seq <= *lastClientSeq {
			return "", ports.ErrInvalidProviderInput
		}
		*lastClientSeq = next.Seq
		if next.SessionID != session.ID {
			return "", ports.ErrForbidden
		}
		if next.Type == realtimeClientMessageClientAck {
			if next.AckSeq <= 0 {
				return "", ports.ErrInvalidProviderInput
			}
			continue
		}
		message = next
		break
	}
	planID := strings.TrimSpace(message.PlanID)
	if planID == "" || planID != strings.TrimSpace(expectedPlanID) {
		return "", ports.ErrForbidden
	}

	var eventType realtimeServerMessageType
	var status string
	switch message.Type {
	case realtimeClientMessageActionPlanApprove:
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
		eventType = realtimeServerMessageType(app.RealtimeVoiceEventActionPlanApproved)
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
		outcomeType := realtimeServerMessageType(app.RealtimeVoiceEventActionPlanExecuted)
		outcomeMessage := "The approved change was applied."
		if err != nil {
			if executed.Record.State != actionplan.StateFailed {
				return "", err
			}
			outcomeType = realtimeServerMessageType(app.RealtimeVoiceEventActionPlanFailed)
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
	case realtimeClientMessageActionPlanCancel:
		if message.PhotoAttachmentsSet {
			return "", ports.ErrInvalidProviderInput
		}
		record, err := application.CancelActionPlan(ctx, app.ActionPlanDecisionInput{
			Principal:   session.Principal,
			TenantID:    session.TenantID,
			InventoryID: session.InventoryID,
			PlanID:      planID,
		})
		if err != nil {
			return "", err
		}
		eventType = realtimeServerMessageType(app.RealtimeVoiceEventActionPlanCancelled)
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
