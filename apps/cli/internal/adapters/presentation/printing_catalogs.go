package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

func (o Output) labelTemplates(templates []ports.LabelTemplate) error {
	for _, v := range templates {
		if err := o.details([][2]string{{"Template", v.ID}, {"Version", strconv.FormatUint(uint64(v.Version), 10)}, {"Name", v.Name}, {"Purpose", v.Purpose}, {"Font", v.Font}, {"Glyph coverage", v.GlyphCoverage}, {"Minimum QR module pixels", strconv.FormatInt(v.MinimumQRModulePixels, 10)}, {"Options", strings.Join(v.Options, ", ")}, {"Show reference by default", strconv.FormatBool(v.Defaults.ShowReference)}}); err != nil {
			return err
		}
	}
	return nil
}
func (o Output) printerProfiles(profiles []ports.PrinterProfile) error {
	for _, v := range profiles {
		if err := o.details([][2]string{{"Adapter", v.AdapterID}, {"Name", v.Name}, {"Transport", v.Transport}, {"Platforms", strings.Join(v.SupportedPlatforms, ", ")}, {"Physically verified", strconv.FormatBool(v.PhysicallyVerified)}}); err != nil {
			return err
		}
		for _, m := range v.Media {
			if _, err := fmt.Fprintf(o.Stdout, "  %s  %s  v%d  %d × %d µm  %d DPI\n", strconv.Quote(m.PresetID), strconv.Quote(m.Name), m.Version, m.WidthMicrometers, m.HeightMicrometers, m.ResolutionDpi); err != nil {
				return err
			}
		}
	}
	return nil
}
