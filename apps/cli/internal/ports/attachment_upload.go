package ports

import (
	"context"
	"io"
)

type DirectUpload struct {
	UploadID     string            `json:"uploadId"`
	AttachmentID string            `json:"attachmentId"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	FormFields   map[string]string `json:"formFields"`
	ExpiresAt    string            `json:"expiresAt"`
}
type AttachmentUploads interface {
	CreateAttachment(context.Context, Scope, string, io.Reader) (Result[Attachment], error)
	StartAttachmentUpload(context.Context, Scope, string, io.Reader) (Result[DirectUpload], error)
	CompleteAttachmentUpload(context.Context, Scope, string, string) (Result[Attachment], error)
}
