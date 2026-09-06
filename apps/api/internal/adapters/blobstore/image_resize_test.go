package blobstore

import (
	"image"
	"image/color"
	"testing"
)

func TestCameraResizeWorkingMemory(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 4032, 3024))
	result := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resized := resizeImage(source, 1600)
			if resized.Bounds().Dx() != 1600 || resized.Bounds().Dy() != 1200 {
				b.Fatal("aspect ratio changed")
			}
		}
	})
	// This excludes the caller-owned input and decode; it bounds resize scratch/output.
	if result.AllocedBytesPerOp() > 100*1024*1024 {
		t.Fatalf("resize allocated %d bytes; budget is 100 MiB", result.AllocedBytesPerOp())
	}
}

func TestLargeReductionFiltersFinePatterns(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	for y := 0; y < 1024; y++ {
		for x := 0; x < 1024; x++ {
			value := uint8(0)
			if (x+y)%2 == 0 {
				value = 255
			}
			source.SetRGBA(x, y, color.RGBA{value, value, value, 255})
		}
	}
	result := resizeImage(source, 64)
	for y := 4; y < 60; y++ {
		for x := 4; x < 60; x++ {
			r, g, b, a := result.At(x, y).RGBA()
			if r < 120*257 || r > 135*257 || r != g || g != b || a != 65535 {
				t.Fatal("fine pattern aliases or loses opacity")
			}
		}
	}
}

func TestReductionPreservesAspectAndTransparentColor(t *testing.T) {
	source := image.NewRGBA(image.Rect(4, 8, 1028, 520))
	for y := 8; y < 520; y++ {
		for x := 4; x < 1028; x++ {
			source.SetRGBA(x, y, color.RGBA{80, 40, 20, 128})
		}
	}
	result := resizeImage(source, 64)
	if result.Bounds() != image.Rect(0, 0, 64, 32) {
		t.Fatal("image geometry changed")
	}
	r, g, b, a := result.At(32, 16).RGBA()
	if r/257 < 78 || r/257 > 82 || g/257 < 38 || g/257 > 42 || b/257 < 18 || b/257 > 22 || a/257 < 126 || a/257 > 130 {
		t.Fatal("premultiplied color changed")
	}
}
