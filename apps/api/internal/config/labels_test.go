package config

import "testing"

func TestLabelSettingsRejectUnsafeOriginsAndLimits(t *testing.T) {
	for _, base := range []string{"http://stash.test", "https://user:password@stash.test", "https://stash.test/path?token=secret", "https://stash.test/path#fragment", "https://stash.test/a/../b", "https://stash.test/a%2fb", "https://stash.test//ambiguous"} {
		if _, err := (LabelConfig{BaseURL: base}).Settings(); err == nil {
			t.Errorf("accepted unsafe label base %s", base)
		}
	}
	configured, err := (LabelConfig{PublicWebBaseURL: "https://stash.test/prefix"}).Settings()
	if err != nil || configured.BaseURL != "https://stash.test/prefix" {
		t.Fatal("public web fallback lost prefix")
	}
	for _, cfg := range []LabelConfig{{RenderTTL: "-1s"}, {MaxRenderBytes: "0"}, {MaxPixels: "999999999"}, {CleanupInterval: "invalid"}} {
		if _, err := cfg.Settings(); err == nil {
			t.Fatal("accepted invalid limits")
		}
	}
}
