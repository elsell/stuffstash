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
		return nil, ports.Failure("input", "The CLI cannot read the terminal settings. Use this command's --help to find the required input options.")
	}
	if err := windows.SetConsoleMode(handle, previous|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING|windows.ENABLE_PROCESSED_OUTPUT); err != nil {
		return nil, ports.Failure("input", "This terminal cannot display the picker. Use a terminal with ANSI support. Use this command's --help for other input options.")
	}
	return func() { _ = windows.SetConsoleMode(handle, previous) }, nil
}
