//go:build !windows

package contextfile

import "os"

func privateFile(info os.FileInfo) bool {
	return info != nil && info.Mode().IsRegular() && info.Mode().Perm()&0077 == 0 && owned(info)
}
func secureFile(file *os.File) bool { info, err := file.Stat(); return err == nil && privateFile(info) }
func secureDirectory(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.IsDir() && info.Mode().Perm()&0022 == 0 && owned(info)
}
func prepareDirectory(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return configError()
	}
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0022 != 0 || !owned(info) {
		return configError()
	}
	return nil
}
