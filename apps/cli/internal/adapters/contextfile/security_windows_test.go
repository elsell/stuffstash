//go:build windows

package contextfile

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/app/contexts"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsContextRejectsBroadFileAccess(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "private", "contexts.json")}
	if err := store.Update(context.Background(), func(c *contexts.Config) error { return nil }); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(store.Path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if !secureFile(file) {
		t.Fatal("created context is not private")
	}
	descriptor, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(store.Path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(context.Background()); err == nil {
		t.Fatal("read publicly accessible context")
	}
	if err := store.Update(context.Background(), func(c *contexts.Config) error { return nil }); err == nil {
		t.Fatal("updated publicly accessible context")
	}
}

func TestWindowsContextRejectsBroadDirectoryAccess(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "private", "contexts.json")}
	if err := store.Update(context.Background(), func(c *contexts.Config) error { return nil }); err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ path, sddl string }{
		{store.Path, "D:P(A;;FA;;;" + user.User.Sid.String() + ")"},
		{filepath.Dir(store.Path), "D:P(A;;FA;;;WD)"},
	} {
		descriptor, err := windows.SecurityDescriptorFromString(item.sddl)
		if err != nil {
			t.Fatal(err)
		}
		acl, _, err := descriptor.DACL()
		if err != nil {
			t.Fatal(err)
		}
		if err := windows.SetNamedSecurityInfo(item.path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.Load(context.Background()); err == nil {
		t.Fatal("read context in publicly accessible directory")
	}
	if err := store.Update(context.Background(), func(c *contexts.Config) error { return nil }); err == nil {
		t.Fatal("updated context in publicly accessible directory")
	}
}
