//go:build linux

package usblp

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoveryDistinguishesIdenticalModelsAndRejectsArbitraryPaths(t *testing.T) {
	root := t.TempDir()
	for _, device := range []struct{ port, serial string }{{"1-2", "serial-a"}, {"1-3", "serial-b"}} {
		dir := filepath.Join(root, device.port)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		for name, value := range map[string]string{"idVendor": "04f9", "idProduct": "209b", "serial": device.serial} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	access := Access{Roots: Roots{SysfsUSB: root, Devices: t.TempDir()}, VendorID: "04f9", ProductID: "209b", Model: "Brother QL-800"}
	devices, err := access.Discover(context.Background())
	if err != nil || len(devices) != 2 {
		t.Fatal(devices, err)
	}
	if devices[0].ID == devices[1].ID || devices[0].State != printing.Unavailable || devices[0].Reason != printing.UnsupportedTransport {
		t.Fatal("model-only identity or false readiness")
	}
	devices[0].Path = "/dev/null"
	if _, err := access.OpenTransport(context.Background(), devices[0]); err == nil {
		t.Fatal("accepted caller device path")
	}
}
