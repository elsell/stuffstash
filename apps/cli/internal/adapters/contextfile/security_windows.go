//go:build windows

package contextfile

import (
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"unsafe"
)

// Windows mode bits do not describe access. Check the opened handle's DACL.
func privateFile(info os.FileInfo) bool { return info != nil && info.Mode().IsRegular() }
func secureFile(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && privateFile(info) && privateHandle(file)
}
func secureDirectory(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.IsDir() && privateHandle(file)
}
func privateHandle(file *os.File) bool {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return false
	}
	descriptor, err := windows.GetSecurityInfo(windows.Handle(file.Fd()), windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return false
	}
	owner, _, err := descriptor.Owner()
	if err != nil || owner == nil || !owner.Equals(user.User.Sid) {
		return false
	}
	acl, _, err := descriptor.DACL()
	if err != nil || acl == nil {
		return false
	}
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return false
	}
	allowed := false
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if windows.GetAce(acl, i, &ace) != nil || ace == nil {
			return false
		}
		// Fail closed for object, callback and other unfamiliar ACE forms.
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return false
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !sid.IsValid() || (!sid.Equals(user.User.Sid) && !sid.Equals(system)) {
			return false
		}
		if sid.Equals(user.User.Sid) {
			allowed = true
		}
	}
	return allowed
}
func prepareDirectory(path string) error {
	// The leaf directory is created with private inherited access, never briefly
	// exposed under a permissive parent ACL. Existing ACLs are never rewritten.
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return configError()
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return configError()
	}
	descriptor, err := windows.SecurityDescriptorFromString("O:" + user.User.Sid.String() + "D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")")
	if err != nil {
		return configError()
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return configError()
	}
	attributes := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: descriptor}
	if err = windows.CreateDirectory(name, &attributes); err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return configError()
	}
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return configError()
	}
	file, err := os.Open(path)
	if err != nil {
		return configError()
	}
	defer file.Close()
	if !secureDirectory(file) {
		return configError()
	}
	return nil
}
