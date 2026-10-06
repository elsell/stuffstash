package ports

import (
	"context"
	"io"
)

// BinaryContent owns a download stream. The caller must close Body and treat
// ContentDisposition as metadata, not as an authorized local path.
type BinaryContent struct {
	Body               io.ReadCloser
	ContentType        string
	ContentDisposition string
	ContentLength      int64
}

// BinaryFiles consumes a stream without closing its caller-owned body.
type BinaryFiles interface {
	PublishContent(context.Context, string, BinaryContent) error
}
