package printingprofiles

import (
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	profiles "github.com/stuffstash/stuff-stash/printingprofiles"
)

type Catalog struct{}

func (Catalog) ResolvePrinterMedia(adapterID, presetID string, version uint32) (printing.MediaSnapshot, error) {
	for _, adapter := range profiles.All() {
		if adapter.ID != adapterID {
			continue
		}
		for _, m := range adapter.Media {
			if m.ID != presetID || m.Version != version {
				continue
			}
			return printing.MediaSnapshot{PresetID: m.ID, Version: m.Version, WidthMicrometers: m.WidthMicrometers, HeightMicrometers: m.HeightMicrometers, MarginsMicrometers: printing.Margins{Left: m.Margins.Left, Right: m.Margins.Right, Top: m.Margins.Top, Bottom: m.Margins.Bottom}, ResolutionDPI: m.ResolutionDPI, RasterWidth: m.RasterWidth, RasterHeight: m.RasterHeight, Orientation: printing.Orientation(m.Orientation), ColorMode: printing.ColorMode(m.ColorMode), CutPolicy: printing.CutPolicy(m.CutPolicy), DisplayRotation: m.DisplayRotation}, nil
		}
	}
	return printing.MediaSnapshot{}, errors.New("unsupported printer media preset")
}

func (c Catalog) ListPrinterProfiles() []printing.PrinterProfile {
	out := make([]printing.PrinterProfile, 0)
	for _, descriptor := range profiles.All() {
		p := printing.PrinterProfile{AdapterID: descriptor.ID, Name: descriptor.Model, Transport: descriptor.Transport, SupportedPlatforms: descriptor.Platforms, PhysicallyVerified: descriptor.PhysicallyVerified, Media: make([]printing.MediaProfile, 0, len(descriptor.Media))}
		for _, m := range descriptor.Media {
			snapshot, _ := c.ResolvePrinterMedia(descriptor.ID, m.ID, m.Version)
			p.Media = append(p.Media, printing.MediaProfile{Name: m.Name, Media: snapshot})
		}
		out = append(out, p)
	}
	return out
}
