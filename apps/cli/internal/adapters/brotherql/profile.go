package brotherql

import "github.com/stuffstash/stuff-stash/cli/internal/domain/printing"

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
	return printing.Media{PresetID: "brother-ql800-29x90", Version: 1, WidthMicrometers: 29000, HeightMicrometers: 89800, Margins: printing.Margins{Left: 1524, Right: 1524, Top: 2963, Bottom: 2963}, ResolutionDPI: 300, RasterWidth: rasterWidth, RasterHeight: rasterHeight, Orientation: "feed", ColorMode: "monochrome", CutPolicy: "after_label", DisplayRotation: 270}
}
func Descriptor() printing.Descriptor {
	return printing.Descriptor{ID: AdapterID, Model: Model, Platforms: []string{"linux"}, Transport: "usb-usblp", ContractVersions: []int{1}, Formats: []string{"image/png"}, Media: []printing.Media{Media()}, CompletionEvidence: "printing_completed_then_waiting", Wake: false, PhysicallyVerified: false}
}
