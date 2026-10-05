//go:build !windows

package inputfiles

import (
	"os"
	"syscall"
)

// Nonblocking open prevents a path replacement with a FIFO from trapping the
// process before the opened handle can be checked for regular-file status.
func openInput(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
