//go:build windows

package credentials

import (
	"errors"
	"os"
)

func checkOwner(os.FileInfo) error {
	return errors.New("file credentials are supported on Unix only; use the OS credential store on Windows")
}
