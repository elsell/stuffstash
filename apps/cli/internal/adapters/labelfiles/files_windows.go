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
	return ports.Failure("unsupported", "saving label files requires Linux or macOS; private Windows file output is not yet supported")
}
