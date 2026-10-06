//go:build windows

package binaryfiles

import (
	"crypto/rand"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"unsafe"
)

func createPrivate(path string) (*os.File, func() error, error) {
	fail := func() (*os.File, func() error, error) {
		return nil, nil, ports.Failure("file", "Cannot create a private file. Check the output directory and permissions.")
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fail()
	}
	sd, err := windows.SecurityDescriptorFromString("O:" + user.User.Sid.String() + "D:P(A;;FA;;;" + user.User.Sid.String() + ")")
	if err != nil {
		return fail()
	}
	name := filepath.Join(filepath.Dir(path), ".stuffstash-download-"+rand.Text())
	ptr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return fail()
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	h, err := windows.CreateFile(ptr, windows.GENERIC_READ|windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, &sa, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return fail()
	}
	f := os.NewFile(uintptr(h), name)
	return f, func() error {
		if err := os.Link(name, path); err != nil {
			return ports.Failure("file", "Cannot publish the file. Choose a new output path on a filesystem that supports hard links.")
		}
		return nil
	}, nil
}
