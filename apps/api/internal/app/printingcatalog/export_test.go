package printingcatalog_test

import (
	"bytes"
	"context"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	"image/png"
	"reflect"
	"sort"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/labelrenderer"
	"github.com/stuffstash/stuff-stash/internal/app/printingcatalog"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func TestExportUsesRuntimeRegistryAndStableArtifacts(t *testing.T) {
	renderer, _ := labelrenderer.New(labelrenderer.DefaultLimits())
	media, err := (printingprofiles.Catalog{}).ResolvePrinterMedia("brother-ql800", "brother-ql800-29x90", 1)
	if err != nil {
		t.Fatal(err)
	}
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
		for _, name := range []string{fixture.PNG, fixture.DisplayPNG} {
			image, err := png.Decode(bytes.NewReader(first.Files[name]))
			if err != nil {
				t.Fatal(err)
			}
			width, height := fixture.WidthPixels, fixture.HeightPixels
			if name == fixture.DisplayPNG {
				width, height = fixture.DisplayWidthPixels, fixture.DisplayHeightPixels
			}
			if image.Bounds().Dx() != width || image.Bounds().Dy() != height {
				t.Fatal("Wrong catalog fixture dimensions")
			}
			bitmap, err := gozxing.NewBinaryBitmapFromImage(image)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := qrcode.NewQRCodeReader().Decode(bitmap, nil)
			if err != nil || decoded.GetText() != printingcatalog.FixtureURL {
				t.Fatalf("Invalid documentation QR %s: %v", name, err)
			}
		}
		if len(first.Files[fixture.PNG]) == 0 || len(first.Files[fixture.DisplayPNG]) == 0 || len(first.Files[fixture.PDF]) == 0 {
			t.Fatal("missing fixture")
		}
	}
	media.RasterWidth = 1
	if _, err := printingcatalog.Export(context.Background(), renderer, renderer, []printing.MediaSnapshot{media}); err == nil {
		t.Fatal("silently swallowed advertised fixture render failure")
	}
}
