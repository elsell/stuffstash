//go:build windows

package terminal

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/sys/windows"
	"os"
)

func prepareOutput(file *os.File) (func(), error) {
	handle := windows.Handle(file.Fd())
	var previous uint32
	if err := windows.GetConsoleMode(handle, &previous); err != nil {
		return nil, ports.Failure("input", "Cannot read the terminal settings. Supply the required scope options.")
	}
	if err := windows.SetConsoleMode(handle, previous|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING|windows.ENABLE_PROCESSED_OUTPUT); err != nil {
		return nil, ports.Failure("input", "This terminal cannot display the picker. Use a terminal with ANSI support or supply scope options.")
	}
	return func() { _ = windows.SetConsoleMode(handle, previous) }, nil
}
