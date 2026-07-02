package httpserver

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func realJPEGAttachmentContent(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 1200, 800))
	for y := range 800 {
		for x := range 1200 {
			img.Set(x, y, color.RGBA{
				R: uint8((x*17 + y*3) % 255),
				G: uint8((x*5 + y*19) % 255),
				B: uint8((x*y + x + y) % 255),
				A: 255,
			})
		}
	}
	buffer := bytes.Buffer{}
	if err := jpeg.Encode(&buffer, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode jpeg fixture: %v", err)
	}
	return buffer.Bytes()
}
