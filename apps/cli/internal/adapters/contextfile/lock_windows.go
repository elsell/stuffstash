//go:build windows

package contextfile

import (
	"context"
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"time"
)

func lock(ctx context.Context, root *os.Root, name string) (func(), error) {
	if info, err := root.Lstat(name); err == nil && !privateFile(info) {
		return nil, configError()
	} else if err != nil && !os.IsNotExist(err) {
		return nil, configError()
	}
	file, err := root.OpenFile(name, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, configError()
	}
	if !secureFile(file) {
		file.Close()
		return nil, configError()
	}
	overlapped := new(windows.Overlapped)
	for {
		if err := ctx.Err(); err != nil {
			file.Close()
			return nil, err
		}
		err = windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, overlapped)
		if err == nil {
			return func() { _ = windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, overlapped); _ = file.Close() }, nil
		}
		if !errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			file.Close()
			return nil, configError()
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			file.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
