package ports

import (
	"context"
)

// AttachmentContent is the shared binary download contract.
type AttachmentContent = BinaryContent
type AttachmentDownloads interface {
	AttachmentContent(context.Context, Scope, string, string) (AttachmentContent, error)
	AttachmentThumbnail(context.Context, Scope, string, string, string) (AttachmentContent, error)
}
