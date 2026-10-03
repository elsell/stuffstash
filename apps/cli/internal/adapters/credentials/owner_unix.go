//go:build !windows

package credentials

import (
	"errors"
	"os"
	"syscall"
)

func checkOwner(info os.FileInfo) error {
	s, ok := info.Sys().(*syscall.Stat_t)
	if !ok || s.Uid != uint32(os.Geteuid()) {
		return errors.New("credential storage must belong to the current user")
	}
	return nil
}
