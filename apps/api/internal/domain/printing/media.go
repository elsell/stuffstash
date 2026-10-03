package printing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type Orientation string
type ColorMode string
type CutPolicy string

const (
	OrientationFeed Orientation = "feed"
	ColorMonochrome ColorMode   = "monochrome"
	CutAfterLabel   CutPolicy   = "after_label"
	CutNone         CutPolicy   = "none"
)

type Margins struct {
	Left   int `json:"left"`
	Right  int `json:"right"`
	Top    int `json:"top"`
	Bottom int `json:"bottom"`
}

// MediaSnapshot describes the device-feed coordinate system. Raster dimensions
// describe the printable area, not the full stock. Rotation is clockwise from
// the feed raster into the human-readable display orientation.
type MediaSnapshot struct {
	PresetID           string      `json:"preset_id"`
	Version            uint32      `json:"version"`
	WidthMicrometers   int         `json:"width_micrometers"`
	HeightMicrometers  int         `json:"height_micrometers"`
	MarginsMicrometers Margins     `json:"margins_micrometers"`
	ResolutionDPI      int         `json:"resolution_dpi"`
	RasterWidth        int         `json:"raster_width"`
	RasterHeight       int         `json:"raster_height"`
	Orientation        Orientation `json:"orientation"`
	ColorMode          ColorMode   `json:"color_mode"`
	CutPolicy          CutPolicy   `json:"cut_policy"`
	DisplayRotation    int         `json:"display_rotation"`
}

func (m MediaSnapshot) Validate(maxPixels int) error {
	if m.WidthMicrometers <= 0 || m.HeightMicrometers <= 0 || m.WidthMicrometers > 1000000 || m.HeightMicrometers > 1000000 {
		return errors.New("label dimensions must be positive and at most one meter")
	}
	p := m.MarginsMicrometers
	if p.Left < 0 || p.Right < 0 || p.Top < 0 || p.Bottom < 0 || p.Left >= m.WidthMicrometers || p.Right >= m.WidthMicrometers-p.Left || p.Top >= m.HeightMicrometers || p.Bottom >= m.HeightMicrometers-p.Top {
		return errors.New("label margins leave no printable area")
	}
	if maxPixels <= 0 || m.RasterWidth <= 0 || m.RasterHeight <= 0 || m.RasterWidth > maxPixels/m.RasterHeight || m.ResolutionDPI <= 0 || m.ResolutionDPI > 2400 {
		return errors.New("invalid label raster dimensions or resolution")
	}

	// The stock's nominal dimensions may round a dot differently from its raster.
	// More than one dot of error would scale QR modules or misstate resolution.
	physical := []int{m.WidthMicrometers - p.Left - p.Right, m.HeightMicrometers - p.Top - p.Bottom}
	raster := []int{m.RasterWidth, m.RasterHeight}
	for i, extent := range physical {
		difference := int64(extent)*int64(m.ResolutionDPI) - int64(raster[i])*25400
		if difference < -25400 || difference > 25400 {
			return errors.New("label printable dimensions do not match raster resolution")
		}
	}
	if m.Orientation != OrientationFeed || m.ColorMode != ColorMonochrome || (m.CutPolicy != CutAfterLabel && m.CutPolicy != CutNone) {
		return errors.New("unsupported label orientation, color mode, or cut policy")
	}
	switch m.DisplayRotation {
	case 0, 90, 180, 270:
	default:
		return errors.New("unsupported display rotation")
	}
	return nil
}

// Fingerprint deliberately excludes catalog identity and revision: configuring
// the same effective media again must make existing waiting jobs eligible.
func (m MediaSnapshot) Fingerprint() string {
	m.PresetID = ""
	m.Version = 0
	data, _ := json.Marshal(m)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
