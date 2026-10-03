package labelrenderer

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"io"

	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

// encodePDF has one fixed page and image, no executable/text/font content. The
// device raster is embedded unchanged; physical stock and margins set placement.
func encodePDF(w io.Writer, raster *image.Gray, media printing.MediaSnapshot) error {
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(raster.Pix); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	points := func(micrometers int) float64 { return float64(micrometers) * 72 / 25400 }
	pageW, pageH := points(media.WidthMicrometers), points(media.HeightMicrometers)
	margin := media.MarginsMicrometers
	imageW := points(media.WidthMicrometers - margin.Left - margin.Right)
	imageH := points(media.HeightMicrometers - margin.Top - margin.Bottom)
	command := []byte(fmt.Sprintf("q\n%.6f 0 0 %.6f %.6f %.6f cm\n/Label Do\nQ\n", imageW, imageH, points(margin.Left), points(margin.Bottom)))
	objects := [][]byte{
		[]byte("<< /Type /Catalog /Pages 2 0 R >>"),
		[]byte("<< /Type /Pages /Kids [3 0 R] /Count 1 >>"),
		[]byte(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.6f %.6f] /Resources << /XObject << /Label 4 0 R >> >> /Contents 5 0 R >>", pageW, pageH)),
		stream([]byte(fmt.Sprintf("/Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceGray /BitsPerComponent 8 /Interpolate false /Filter /FlateDecode", raster.Bounds().Dx(), raster.Bounds().Dy())), compressed.Bytes()),
		stream(nil, command),
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, len(objects)+1)
	for i, object := range objects {
		offsets[i+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n", i+1)
		out.Write(object)
		out.WriteString("\nendobj\n")
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	_, err := w.Write(out.Bytes())
	return err
}
func stream(attributes, data []byte) []byte {
	var out bytes.Buffer
	fmt.Fprintf(&out, "<< %s /Length %d >>\nstream\n", attributes, len(data))
	out.Write(data)
	out.WriteString("\nendstream")
	return out.Bytes()
}
