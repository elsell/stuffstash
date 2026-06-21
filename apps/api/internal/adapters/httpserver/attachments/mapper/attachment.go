package mapper

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/attachments/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func AttachmentToResponse(attachment media.Attachment) dto.AttachmentResponse {
	return dto.AttachmentResponse{
		ID:             attachment.ID.String(),
		TenantID:       attachment.TenantID.String(),
		InventoryID:    attachment.InventoryID.String(),
		AssetID:        attachment.AssetID.String(),
		FileName:       attachment.FileName.String(),
		ContentType:    attachment.ContentType.String(),
		SizeBytes:      attachment.SizeBytes,
		SHA256:         attachment.SHA256.String(),
		CreatedAt:      attachment.CreatedAt.UTC().Format(time.RFC3339),
		LifecycleState: attachment.LifecycleState.String(),
	}
}

func AttachmentsToResponse(items []media.Attachment) []dto.AttachmentResponse {
	data := make([]dto.AttachmentResponse, 0, len(items))
	for _, item := range items {
		data = append(data, AttachmentToResponse(item))
	}
	return data
}

func DirectUploadToResponse(upload ports.DirectAttachmentUpload) dto.DirectUploadResponse {
	headers := map[string]string{}
	for key, value := range upload.Headers {
		headers[key] = value
	}
	formFields := map[string]string{}
	for key, value := range upload.FormFields {
		formFields[key] = value
	}
	return dto.DirectUploadResponse{
		UploadID:     upload.UploadID,
		AttachmentID: upload.AttachmentID.String(),
		Method:       upload.Method,
		URL:          upload.URL,
		Headers:      headers,
		FormFields:   formFields,
		ExpiresAt:    upload.ExpiresAt.UTC().Format(time.RFC3339),
	}
}
