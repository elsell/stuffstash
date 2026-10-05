package ports

import (
	"context"
	"io"
)

// AttachmentContent owns a stream. The caller must close Body.
type AttachmentContent struct {
	Body               io.ReadCloser
	ContentType        string
	ContentDisposition string
	ContentLength      int64
}
type AttachmentDownloads interface {
	AttachmentContent(context.Context, Scope, string, string) (AttachmentContent, error)
	AttachmentThumbnail(context.Context, Scope, string, string, string) (AttachmentContent, error)
}
