package ports

import (
	"context"
	"io"
)

// StreamFiles opens a caller-owned binary source without a JSON size limit.
type StreamFiles interface {
	OpenStream(context.Context, string) (io.ReadCloser, error)
}
