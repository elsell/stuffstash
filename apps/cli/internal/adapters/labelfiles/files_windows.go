//go:build windows

package labelfiles

import (
	"context"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Files struct{}

// Unix file modes do not establish a private Windows DACL. Refuse publication
// until an adapter can guarantee that privacy before any label bytes are written.
func (Files) Publish(context.Context, string, []byte) error {
	return ports.Failure("unsupported", "The CLI can save label files on Linux or macOS only. Use Linux or macOS to save the label file.")
}
