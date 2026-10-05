package httpapi

import (
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func printerMedia(v generated.MediaProfile) ports.PrinterMedia {
	return ports.PrinterMedia{Name: v.Name, PresetID: v.PresetId, Version: uint32(v.Version), ColorMode: v.ColorMode, CutPolicy: v.CutPolicy, DisplayRotation: v.DisplayRotation, HeightMicrometers: v.HeightMicrometers, WidthMicrometers: v.WidthMicrometers, Orientation: v.Orientation, RasterHeight: v.RasterHeight, RasterWidth: v.RasterWidth, ResolutionDpi: v.ResolutionDpi, MarginsMicrometers: ports.PrinterMargins{Top: v.MarginsMicrometers.Top, Bottom: v.MarginsMicrometers.Bottom, Left: v.MarginsMicrometers.Left, Right: v.MarginsMicrometers.Right}}
}
