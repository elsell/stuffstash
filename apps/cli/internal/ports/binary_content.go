package ports

import "io"

// BinaryContent owns a download stream. The caller must close Body and treat
// ContentDisposition as metadata, not as an authorized local path.
type BinaryContent struct {
	Body               io.ReadCloser
	ContentType        string
	ContentDisposition string
	ContentLength      int64
}
