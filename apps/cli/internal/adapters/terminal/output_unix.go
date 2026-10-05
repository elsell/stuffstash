//go:build !windows

package terminal

import "os"

func prepareOutput(*os.File) (func(), error) { return func() {}, nil }
