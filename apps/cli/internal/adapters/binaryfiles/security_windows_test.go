//go:build windows

package binaryfiles

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/sys/windows"
)

func TestWindowsPublishedDownloadHasPrivateOwnerAndDACL(t *testing.T) {
	directory := t.TempDir()
	// A permissive parent must not grant access to a newly downloaded private file.
	parentDescriptor, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	parentACL, _, err := parentDescriptor.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err = windows.SetNamedSecurityInfo(directory, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, parentACL, nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "download.bin")
	const payload = "private inventory archive"
	if err = (Files{}).PublishContent(context.Background(), path, ports.BinaryContent{Body: io.NopCloser(strings.NewReader(payload)), ContentLength: int64(len(payload))}); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	descriptor, err := windows.GetSecurityInfo(windows.Handle(file.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || !owner.Equals(user.User.Sid) {
		t.Fatalf("download is not owned by the current user: %v", err)
	}
	control, _, err := descriptor.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatalf("download DACL is not protected: %v", err)
	}
	acl, _, err := descriptor.DACL()
	if err != nil || acl == nil || acl.AceCount == 0 {
		t.Fatalf("download has no restrictive access list: %v", err)
	}
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil || ace == nil {
			t.Fatalf("cannot inspect download grant: %v", err)
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags&windows.INHERITED_ACE != 0 {
			t.Fatal("download contains an unexpected or inherited grant")
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !sid.IsValid() || !sid.Equals(user.User.Sid) {
			t.Fatal("download grants access to another principal")
		}
	}
	content, err := io.ReadAll(file)
	if err != nil || string(content) != payload {
		t.Fatalf("published content changed: %v", err)
	}
}
