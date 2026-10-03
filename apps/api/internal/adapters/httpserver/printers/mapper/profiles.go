package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func Profiles(profiles []printing.PrinterProfile) []dto.PrinterProfile {
	out := make([]dto.PrinterProfile, 0, len(profiles))
	for _, p := range profiles {
		profile := dto.PrinterProfile{AdapterID: p.AdapterID, Name: p.Name, Transport: p.Transport, SupportedPlatforms: p.SupportedPlatforms, PhysicallyVerified: p.PhysicallyVerified, Media: make([]dto.MediaProfile, 0, len(p.Media))}
		for _, named := range p.Media {
			profile.Media = append(profile.Media, Media(named))
		}
		out = append(out, profile)
	}
	return out
}

func Media(named printing.MediaProfile) dto.MediaProfile {
	m := named.Media
	return dto.MediaProfile{Name: named.Name, PresetID: m.PresetID, Version: m.Version, WidthMicrometers: m.WidthMicrometers, HeightMicrometers: m.HeightMicrometers, MarginsMicrometers: dto.MediaMargins{Left: m.MarginsMicrometers.Left, Right: m.MarginsMicrometers.Right, Top: m.MarginsMicrometers.Top, Bottom: m.MarginsMicrometers.Bottom}, ResolutionDPI: m.ResolutionDPI, RasterWidth: m.RasterWidth, RasterHeight: m.RasterHeight, Orientation: string(m.Orientation), ColorMode: string(m.ColorMode), CutPolicy: string(m.CutPolicy), DisplayRotation: m.DisplayRotation}
}
