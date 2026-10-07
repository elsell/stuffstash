package usblp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Roots are environment-backed bootstrap settings; tests supply a faithful sysfs
// tree. Discovery reads metadata only and never opens a USB character device.
type Roots struct{ SysfsUSB, Devices string }
type Access struct {
	Roots                      Roots
	VendorID, ProductID, Model string
}

func (a Access) Discover(ctx context.Context) ([]printing.Device, error) {
	if runtime.GOOS != "linux" {
		return nil, ports.Failure("unsupported", "USB printer discovery requires Linux. Run stuffstash printers discover on Linux.")
	}
	entries, err := os.ReadDir(a.Roots.SysfsUSB)
	if os.IsNotExist(err) {
		return []printing.Device{}, nil
	}
	if err != nil {
		return nil, err
	}
	devices := make([]printing.Device, 0)
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		base := filepath.Join(a.Roots.SysfsUSB, entry.Name())
		if read(base, "idVendor") != a.VendorID || read(base, "idProduct") != a.ProductID {
			continue
		}
		serial := read(base, "serial")
		identity := a.VendorID + ":" + a.ProductID + ":serial:" + serial
		if serial == "" {
			identity = a.VendorID + ":" + a.ProductID + ":port:" + entry.Name()
		}
		digest := sha256.Sum256([]byte(identity))
		device := printing.Device{ID: "usb-" + hex.EncodeToString(digest[:]), Model: a.Model, Serial: serial, PhysicalPort: entry.Name(), State: printing.Unavailable, Reason: printing.UnsupportedTransport}
		interfaces, _ := filepath.Glob(base + ":*")
		for _, iface := range interfaces {
			driver, err := filepath.EvalSymlinks(filepath.Join(iface, "driver"))
			if err != nil || filepath.Base(driver) != "usblp" || read(iface, "bInterfaceProtocol") != "02" {
				continue
			}
			nodes, _ := filepath.Glob(filepath.Join(iface, "usbmisc", "lp*"))
			for _, node := range nodes {
				path := filepath.Join(a.Roots.Devices, filepath.Base(node))
				info, err := os.Lstat(path)
				if err != nil {
					if os.IsPermission(err) {
						device.Reason = printing.PermissionDenied
					}
					continue
				}
				if info.Mode()&os.ModeCharDevice == 0 {
					continue
				}
				device.Path = path
				device.State = printing.Unknown
				device.Reason = printing.NoReason
			}
		}
		devices = append(devices, device)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].ID < devices[j].ID })
	return devices, nil
}
func read(root, name string) string {
	value, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(value))
}
