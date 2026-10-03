// Package printingprofiles owns public, deterministic built-in device metadata.
// It has no transport dependencies and performs no hardware discovery.
package printingprofiles

type Margins struct{ Left, Right, Top, Bottom int }
type Media struct {
	ID                                       string
	Version                                  uint32
	Name                                     string
	WidthMicrometers, HeightMicrometers      int
	Margins                                  Margins
	ResolutionDPI, RasterWidth, RasterHeight int
	Orientation, ColorMode, CutPolicy        string
	DisplayRotation                          int
}
type Descriptor struct {
	ID, Model, Transport string
	Platforms            []string
	Media                []Media
	PhysicallyVerified   bool
}

func BrotherQL800() Descriptor {
	return Descriptor{ID: "brother-ql800", Model: "Brother QL-800", Transport: "usb-usblp", Platforms: []string{"linux"}, Media: []Media{{ID: "brother-ql800-29x90", Version: 1, Name: "29 × 90 mm (DK-11201)", WidthMicrometers: 29000, HeightMicrometers: 89800, Margins: Margins{Left: 1524, Right: 1524, Top: 2963, Bottom: 2963}, ResolutionDPI: 300, RasterWidth: 306, RasterHeight: 991, Orientation: "feed", ColorMode: "monochrome", CutPolicy: "after_label", DisplayRotation: 270}}, PhysicallyVerified: false}
}
func All() []Descriptor { return []Descriptor{BrotherQL800()} }
