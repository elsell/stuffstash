//go:build windows

package credentials

import (
	"os"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func checkOwner(os.FileInfo) error {
	return ports.Failure("configuration", "The CLI supports file credentials on Unix only. Use the OS credential store on Windows.")
}
