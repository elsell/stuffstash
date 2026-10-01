package app

import (
	"context"

	mediaapp "github.com/stuffstash/stuff-stash/internal/app/media"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type InitiateAttachmentDirectUploadInput = mediaapp.InitiateAttachmentDirectUploadInput
type CompleteAttachmentDirectUploadInput = mediaapp.CompleteAttachmentDirectUploadInput

func (a App) InitiateAttachmentDirectUpload(ctx context.Context, input InitiateAttachmentDirectUploadInput) (ports.DirectAttachmentUpload, error) {
	return a.attachmentService().InitiateAttachmentDirectUpload(ctx, input)
}
func (a App) CompleteAttachmentDirectUpload(ctx context.Context, input CompleteAttachmentDirectUploadInput) (media.Attachment, error) {
	return a.attachmentService().CompleteAttachmentDirectUpload(ctx, input)
}
