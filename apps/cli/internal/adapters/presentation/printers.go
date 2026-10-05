package presentation

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"time"
)

func (o Output) registeredPrinter(p ports.RegisteredPrinter) error {
	fields := [][2]string{{"Printer", p.ID}, {"Name", p.Name}, {"Adapter", p.AdapterID}, {"Readiness", p.Readiness}, {"Retired", strconv.FormatBool(p.Retired)}, {"Revision", strconv.FormatUint(p.Revision, 10)}, {"Media fingerprint", p.MediaFingerprint}}
	if p.ReadinessReason != nil {
		fields = append(fields, [2]string{"Readiness reason", *p.ReadinessReason})
	}
	if p.ReportedAt != nil {
		fields = append(fields, [2]string{"Reported", p.ReportedAt.Format(time.RFC3339Nano)})
	}
	if err := o.details(fields); err != nil {
		return err
	}
	m := p.Media
	fields = [][2]string{{"Media", m.Name}, {"Preset", m.PresetID}, {"Orientation", m.Orientation}, {"Color", m.ColorMode}, {"Cut policy", m.CutPolicy}}
	for _, v := range []struct {
		label string
		value int64
	}{{"Version", int64(m.Version)}, {"Width (micrometers)", m.WidthMicrometers}, {"Height (micrometers)", m.HeightMicrometers}, {"Top margin", m.MarginsMicrometers.Top}, {"Bottom margin", m.MarginsMicrometers.Bottom}, {"Left margin", m.MarginsMicrometers.Left}, {"Right margin", m.MarginsMicrometers.Right}, {"Raster width", m.RasterWidth}, {"Raster height", m.RasterHeight}, {"Resolution (DPI)", m.ResolutionDpi}, {"Rotation", m.DisplayRotation}} {
		fields = append(fields, [2]string{v.label, strconv.FormatInt(v.value, 10)})
	}
	return o.details(fields)
}
