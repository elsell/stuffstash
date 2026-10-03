package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"runtime"
	"testing"
)

func TestVersionReportsCompiledUSBSupportWithoutProbingHardware(t *testing.T) {
	var stdout, stderr bytes.Buffer
	// No server/login or accessible device tree is needed to describe the binary.
	getenv := func(key string) string {
		if key == "STUFF_STASH_CLI_SYSFS_USB_ROOT" {
			return "/not-a-device-tree"
		}
		return ""
	}
	if code := Run(context.Background(), []string{"version", "--json"}, getenv, &stdout, &stderr); code != 0 {
		t.Fatal(code, stderr.String())
	}
	var result struct {
		OS          string `json:"os"`
		USBPrinting bool   `json:"usbPrinting"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.OS != runtime.GOOS || result.USBPrinting != (runtime.GOOS == "linux") {
		t.Fatalf("incorrect compiled capability: %s", stdout.String())
	}
}
