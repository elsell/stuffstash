package labelrenderer_test

import (
	"bytes"
	"context"
	"image/png"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/stuffstash/stuff-stash/internal/adapters/labelrenderer"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func testMedia() printing.MediaSnapshot {
	return printing.MediaSnapshot{PresetID: "brother-ql800-29x90", Version: 1, WidthMicrometers: 29000, HeightMicrometers: 90000, MarginsMicrometers: printing.Margins{Left: 1546, Right: 1546, Top: 3047, Bottom: 3048}, ResolutionDPI: 300, RasterWidth: 306, RasterHeight: 991, Orientation: printing.OrientationFeed, ColorMode: printing.ColorMonochrome, CutPolicy: printing.CutAfterLabel, DisplayRotation: 270}
}
func request(id printing.TemplateID, ref bool) printing.RenderRequest {
	return printing.RenderRequest{Content: printing.ContentSnapshot{QRURL: "https://example.invalid/l/v1/01JEXAMPLEINSTANCE/01JEXAMPLELABEL", Title: "Café tools — Καλημέρα — Инструменты", Reference: "STASH-012345"}, Template: printing.TemplateSelection{ID: id, Version: 1, Options: printing.TemplateOptions{ShowReference: ref}}, Media: testMedia(), Format: printing.FormatPNG}
}
func TestTemplatesProduceDeterministicScannableLabels(t *testing.T) {
	r, err := labelrenderer.New(labelrenderer.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []printing.TemplateID{printing.TemplateQRTitle, printing.TemplateQROnly} {
		for _, ref := range []bool{true, false} {
			t.Run(string(id), func(t *testing.T) {
				req := request(id, ref)
				first, err := r.Render(context.Background(), req)
				if err != nil {
					t.Fatal(err)
				}
				second, err := r.Render(context.Background(), req)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(first.Content, second.Content) || first.SHA256 != second.SHA256 {
					t.Fatal("render is not deterministic")
				}
				img, err := png.Decode(bytes.NewReader(first.Content))
				if err != nil {
					t.Fatal(err)
				}
				if img.Bounds().Dx() != 306 || img.Bounds().Dy() != 991 {
					t.Fatal(img.Bounds())
				}
				bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
				if err != nil {
					t.Fatal(err)
				}
				decoded, err := qrcode.NewQRCodeReader().Decode(bitmap, nil)
				if err != nil {
					t.Fatal(err)
				}
				if decoded.GetText() != req.Content.QRURL {
					t.Fatalf("wrong QR URL: %s", decoded.GetText())
				}
				display, err := r.DisplayPNG(first)
				if err != nil {
					t.Fatal(err)
				}
				displayed, err := png.Decode(bytes.NewReader(display))
				if err != nil {
					t.Fatal(err)
				}
				if displayed.Bounds().Dx() != 991 || displayed.Bounds().Dy() != 306 {
					t.Fatal("wrong display orientation")
				}
				displayBitmap, _ := gozxing.NewBinaryBitmapFromImage(displayed)
				displayDecoded, err := qrcode.NewQRCodeReader().Decode(displayBitmap, nil)
				if err != nil || displayDecoded.GetText() != req.Content.QRURL {
					t.Fatalf("display transform damaged QR: %v", err)
				}
			})
		}
	}
}
func TestRejectsUnprintableContentAndProfiles(t *testing.T) {
	r, err := labelrenderer.New(labelrenderer.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*printing.RenderRequest){
		"unsupported glyph": func(r *printing.RenderRequest) { r.Content.Title = "箱" },
		"untrusted URL":     func(r *printing.RenderRequest) { r.Content.QRURL = "http://example.invalid/label" },
		"dense QR": func(r *printing.RenderRequest) {
			r.Content.QRURL = "https://example.invalid/" + string(bytes.Repeat([]byte("x"), 1500))
		},
		"inconsistent DPI": func(r *printing.RenderRequest) { r.Media.ResolutionDPI = 600 },
		"negative margin":  func(r *printing.RenderRequest) { r.Media.MarginsMicrometers.Left = -1 },
		"unknown version":  func(r *printing.RenderRequest) { r.Template.Version = 42 },
		"oversized raster": func(r *printing.RenderRequest) { r.Media.RasterWidth = 100000 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := request(printing.TemplateQRTitle, true)
			mutate(&req)
			if _, err := r.Render(context.Background(), req); err == nil {
				t.Fatal("accepted invalid render")
			}
		})
	}
}
func TestLongTitleTruncatesWithoutDamagingQR(t *testing.T) {
	r, _ := labelrenderer.New(labelrenderer.DefaultLimits())
	req := request(printing.TemplateQRTitle, true)
	req.Content.Title = string(bytes.Repeat([]byte("Café tools "), 80))
	rendered, err := r.Render(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(rendered.Content))
	if err != nil {
		t.Fatal(err)
	}
	bitmap, _ := gozxing.NewBinaryBitmapFromImage(img)
	decoded, err := qrcode.NewQRCodeReader().Decode(bitmap, nil)
	if err != nil || decoded.GetText() != req.Content.QRURL {
		t.Fatalf("long title damaged QR: %v", err)
	}
}
