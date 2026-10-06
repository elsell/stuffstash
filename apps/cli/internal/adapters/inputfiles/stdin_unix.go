//go:build !windows

package inputfiles

import (
	"golang.org/x/sys/unix"
	"io"
	"os"
)

// Inherited stdin can be a blocking descriptor that Go cannot interrupt by
// closing it. Give the reader a nonblocking duplicate before os.NewFile so Go
// registers the duplicate with its poller. Restore shared descriptor flags when
// the finite command finishes.
func prepareStdin(reader io.Reader) (io.Reader, func(), error) {
	original, ok := reader.(*os.File)
	if !ok {
		return reader, func() {}, nil
	}
	flags, err := unix.FcntlInt(original.Fd(), unix.F_GETFL, 0)
	if err != nil {
		return nil, nil, err
	}
	fd, err := unix.Dup(int(original.Fd()))
	if err != nil {
		return nil, nil, err
	}
	unix.CloseOnExec(fd)
	if err := unix.SetNonblock(fd, true); err != nil {
		unix.Close(fd)
		return nil, nil, err
	}
	file := os.NewFile(uintptr(fd), "request-stdin")
	cleanup := func() { file.Close(); _, _ = unix.FcntlInt(original.Fd(), unix.F_SETFL, flags) }
	return file, cleanup, nil
}
