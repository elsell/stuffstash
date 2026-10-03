package labelrenderer_test

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/stuffstash/stuff-stash/internal/adapters/labelrenderer"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func TestPDFPhysicalSizeAndIndependentRasterDecoding(t *testing.T) {
	for _, tool := range []string{"pdfinfo", "pdftoppm"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip("independent PDF verification requires Poppler: " + tool)
		}
	}
	renderer, _ := labelrenderer.New(labelrenderer.DefaultLimits())
	req := request(printing.TemplateQRTitle, true)
	req.Format = printing.FormatPDF
	rendered, err := renderer.Render(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	again, err := renderer.Render(context.Background(), req)
	if err != nil || !bytes.Equal(rendered.Content, again.Content) {
		t.Fatal("PDF is not deterministic")
	}
	dir := t.TempDir()
	pdf := filepath.Join(dir, "label.pdf")
	if err := os.WriteFile(pdf, rendered.Content, 0600); err != nil {
		t.Fatal(err)
	}
	info, err := exec.Command("pdfinfo", pdf).CombinedOutput()
	if err != nil {
		t.Fatalf("invalid PDF: %v: %s", err, info)
	}
	// 29 x 90 mm, expressed independently in PDF points by Poppler.
	if !strings.Contains(string(info), "82.2047 x 255.118 pts") {
		t.Fatalf("unexpected physical size: %s", info)
	}
	output := filepath.Join(dir, "page")
	if data, err := exec.Command("pdftoppm", "-singlefile", "-r", "300", "-png", pdf, output).CombinedOutput(); err != nil {
		t.Fatalf("render PDF: %v: %s", err, data)
	}
	file, err := os.Open(output + ".png")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal("PDF QR changed")
	}
}
