package version

import "runtime"

// Release automation sets these using linker flags. Local builds stay explicit.
var Tag = "development"
var Commit = "unknown"

type Info struct {
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	USBPrinting  bool   `json:"usbPrinting"`
}

func Current() Info {
	return Info{Version: Tag, Commit: Commit, OS: runtime.GOOS, Architecture: runtime.GOARCH, USBPrinting: false}
}
