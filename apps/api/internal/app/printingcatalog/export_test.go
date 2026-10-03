package printingcatalog_test

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/labelrenderer"
	"github.com/stuffstash/stuff-stash/internal/app/printingcatalog"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func TestExportUsesRuntimeRegistryAndStableArtifacts(t *testing.T) {
	renderer, _ := labelrenderer.New(labelrenderer.DefaultLimits())
	media := printing.MediaSnapshot{PresetID: "test-29x90", Version: 1, WidthMicrometers: 29000, HeightMicrometers: 90000, MarginsMicrometers: printing.Margins{Left: 1546, Right: 1546, Top: 3047, Bottom: 3048}, ResolutionDPI: 300, RasterWidth: 306, RasterHeight: 991, Orientation: printing.OrientationFeed, ColorMode: printing.ColorMonochrome, CutPolicy: printing.CutAfterLabel, DisplayRotation: 270}
	first, err := printingcatalog.Export(context.Background(), renderer, renderer, []printing.MediaSnapshot{media})
	if err != nil {
		t.Fatal(err)
	}
	second, err := printingcatalog.Export(context.Background(), renderer, renderer, []printing.MediaSnapshot{media})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("catalog not deterministic")
	}
	runtimeTemplates := renderer.Templates()
	sort.Slice(runtimeTemplates, func(i, j int) bool { return runtimeTemplates[i].ID < runtimeTemplates[j].ID })
	if !reflect.DeepEqual(first.Catalog.Templates, runtimeTemplates) {
		t.Fatal("catalog differs from runtime registry")
	}
	if len(first.Catalog.Examples) != 4 {
		t.Fatal("missing template/options previews")
	}
	for _, fixture := range first.Catalog.Examples {
		if len(first.Files[fixture.PNG]) == 0 || len(first.Files[fixture.DisplayPNG]) == 0 || len(first.Files[fixture.PDF]) == 0 {
			t.Fatal("missing fixture")
		}
	}
	media.RasterWidth = 1
	if _, err := printingcatalog.Export(context.Background(), renderer, renderer, []printing.MediaSnapshot{media}); err == nil {
		t.Fatal("silently swallowed advertised fixture render failure")
	}
}
