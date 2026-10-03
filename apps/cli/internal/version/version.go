package version

import (
	"runtime"
	"strings"
)

// Release automation sets these using linker flags. Local builds stay explicit.
var Build = "development:unknown"

type Info struct {
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	USBPrinting  bool   `json:"usbPrinting"`
}

func Current() Info {
	tag, commit, _ := strings.Cut(Build, ":")
	return Info{Version: tag, Commit: commit, OS: runtime.GOOS, Architecture: runtime.GOARCH, USBPrinting: false}
}
