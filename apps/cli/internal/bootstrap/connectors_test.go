package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/brotherql"
)

func TestPairingReviewCandidateDoesNotExposePhysicalDeviceIdentity(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux sysfs fixture")
	}
	root := t.TempDir()
	device := filepath.Join(root, "1-2")
	if err := os.Mkdir(device, 0700); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{"idVendor": brotherql.VendorID, "idProduct": brotherql.ProductID, "serial": "private-serial"} {
		if err := os.WriteFile(filepath.Join(device, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	candidates, err := discoverPairingCandidates(context.Background(), func(key string) string {
		if key == "STUFF_STASH_CLI_SYSFS_USB_ROOT" {
			return root
		}
		return ""
	})
	if err != nil || len(candidates) != 1 {
		t.Fatalf("discover candidates: %v %v", candidates, err)
	}
	if candidates[0].ID == candidates[0].DeviceID || candidates[0].ID != "candidate-1" {
		t.Fatal("public candidate ID contains stable device identity")
	}
	if candidates[0].DeviceID == "" {
		t.Fatal("protected binding lost device identity")
	}
}
