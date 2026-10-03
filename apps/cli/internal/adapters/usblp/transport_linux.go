//go:build linux

package usblp

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/sys/unix"
	"io"
	"sync"
)

type transport struct {
	fd   int
	once sync.Once
}

func (a Access) OpenTransport(ctx context.Context, device printing.Device) (ports.PrinterTransport, error) {
	// Resolve again from trusted local metadata; callers cannot supply an arbitrary
	// path received from an API job or stale registration.
	devices, err := a.Discover(ctx)
	if err != nil {
		return nil, err
	}
	path := ""
	for _, current := range devices {
		if current.ID == device.ID && current.Path != "" && current.Path == device.Path {
			path = current.Path
			break
		}
	}
	if path == "" {
		return nil, ports.Failure("unavailable", "printer is not available through a bidirectional usblp device")
	}
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_NONBLOCK|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		if errors.Is(err, unix.EACCES) || errors.Is(err, unix.EPERM) {
			return nil, ports.Failure("permission_denied", "grant the printer group access to its USB device")
		}
		return nil, ports.Failure("unavailable", "printer is disconnected or in use")
	}
	var info unix.Stat_t
	if err := unix.Fstat(fd, &info); err != nil || info.Mode&unix.S_IFMT != unix.S_IFCHR {
		unix.Close(fd)
		return nil, ports.Failure("unavailable", "printer path is not a character device")
	}
	return &transport{fd: fd}, nil
}
func (t *transport) Read(ctx context.Context, p []byte) (int, error) {
	return t.transfer(ctx, p, false)
}
func (t *transport) Write(ctx context.Context, p []byte) (int, error) {
	return t.transfer(ctx, p, true)
}
func (t *transport) transfer(ctx context.Context, p []byte, write bool) (int, error) {
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		var n int
		var err error
		if write {
			n, err = unix.Write(t.fd, p)
		} else {
			n, err = unix.Read(t.fd, p)
		}
		if err == nil {
			if n == 0 && !write {
				return 0, io.EOF
			}
			return n, nil
		}
		if n > 0 {
			return n, err
		}
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if !errors.Is(err, unix.EAGAIN) {
			return 0, err
		}
		event := int16(unix.POLLIN)
		if write {
			event = unix.POLLOUT
		}
		fds := []unix.PollFd{{Fd: int32(t.fd), Events: event}}
		_, err = unix.Poll(fds, 100)
		if err != nil && !errors.Is(err, unix.EINTR) {
			return 0, err
		}
		if fds[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return 0, io.ErrUnexpectedEOF
		}
	}
}
func (t *transport) Close() error {
	var err error
	t.once.Do(func() { err = unix.Close(t.fd) })
	return err
}
