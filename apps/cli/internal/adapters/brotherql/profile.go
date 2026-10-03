package brotherql

import (
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/printingprofiles"
)

const (
	AdapterID       = "brother-ql800"
	Model           = "Brother QL-800"
	VendorID        = "04f9"
	ProductID       = "209b"
	rasterWidth     = 306
	rasterHeight    = 991
	rasterBytes     = 90
	rightHeadOffset = 6
)

func Media() printing.Media {
	m := printingprofiles.BrotherQL800().Media[0]
	return printing.Media{PresetID: m.ID, Version: m.Version, WidthMicrometers: m.WidthMicrometers, HeightMicrometers: m.HeightMicrometers, Margins: printing.Margins{Left: m.Margins.Left, Right: m.Margins.Right, Top: m.Margins.Top, Bottom: m.Margins.Bottom}, ResolutionDPI: m.ResolutionDPI, RasterWidth: m.RasterWidth, RasterHeight: m.RasterHeight, Orientation: m.Orientation, ColorMode: m.ColorMode, CutPolicy: m.CutPolicy, DisplayRotation: m.DisplayRotation}
}
func Descriptor() printing.Descriptor {
	p := printingprofiles.BrotherQL800()
	return printing.Descriptor{ID: p.ID, Model: p.Model, Platforms: p.Platforms, Transport: p.Transport, ContractVersions: []int{1}, Formats: []string{"image/png"}, Media: []printing.Media{Media()}, CompletionEvidence: "printing_completed_then_waiting", Wake: false, PhysicallyVerified: p.PhysicallyVerified}
}
