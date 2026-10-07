//go:build !windows

package credentials

import (
	"os"
	"syscall"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func checkOwner(info os.FileInfo) error {
	s, ok := info.Sys().(*syscall.Stat_t)
	if !ok || s.Uid != uint32(os.Geteuid()) {
		return ports.Failure("configuration", "The credential storage must belong to your account. Select a private path that your account owns.")
	}
	return nil
}
