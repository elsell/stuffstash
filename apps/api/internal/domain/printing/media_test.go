package printing

import (
	"testing"
)

func TestMediaFingerprintTracksEffectiveSettingsNotCatalogRevision(t *testing.T) {
	original := MediaSnapshot{PresetID: "initial", Version: 1, WidthMicrometers: 29000, HeightMicrometers: 90000, RasterWidth: 306, RasterHeight: 991, ResolutionDPI: 300, Orientation: OrientationFeed, ColorMode: ColorMonochrome, CutPolicy: CutAfterLabel, DisplayRotation: 270}
	updated := original
	updated.PresetID = "renamed"
	updated.Version = 2
	if updated.Fingerprint() != original.Fingerprint() {
		t.Fatal("equivalent media would strand queued jobs")
	}
	updated.RasterWidth++
	if updated.Fingerprint() == original.Fingerprint() {
		t.Fatal("changed raster would silently reuse incompatible jobs")
	}
	updated = original
	updated.DisplayRotation = 90
	if updated.Fingerprint() == original.Fingerprint() {
		t.Fatal("changed orientation must change fingerprint")
	}
}
