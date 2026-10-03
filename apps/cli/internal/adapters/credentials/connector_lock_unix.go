//go:build !windows

package credentials

import (
	"golang.org/x/sys/unix"
	"os"
)

func lockConnectorFile(path string) (func(), error) {
	descriptor, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(descriptor), path)
	if err := checkFile(path); err != nil {
		file.Close()
		return nil, err
	}
	if err := unix.Flock(descriptor, unix.LOCK_EX); err != nil {
		file.Close()
		return nil, err
	}
	return func() { _ = unix.Flock(descriptor, unix.LOCK_UN); _ = file.Close() }, nil
}
