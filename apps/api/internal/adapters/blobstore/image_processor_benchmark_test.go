package blobstore

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// This isolates codec CPU/allocation cost; it does not measure storage or device latency.
func BenchmarkCameraPhotoThumbnail(b *testing.B) {
	const width, height = 4032, 3024
	source := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			source.SetRGBA(x, y, color.RGBA{R: uint8(x*17 + y*3), G: uint8(x*5 + y*19), B: uint8(x*y + x + y), A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, source, &jpeg.Options{Quality: 92}); err != nil {
		b.Fatal(err)
	}
	processor := StandardImageProcessor{}
	for _, variant := range []media.ThumbnailVariant{media.ThumbnailVariantSmall, media.ThumbnailVariantMedium, media.ThumbnailVariantLarge} {
		b.Run(variant.String(), func(b *testing.B) {
			request := ports.ImageDerivativeRequest{Variant: variant, ContentType: media.ContentTypeJPEG, Content: encoded.Bytes()}
			b.ReportAllocs()
			b.SetBytes(int64(encoded.Len()))
			b.ResetTimer()
			var size int
			for i := 0; i < b.N; i++ {
				result, err := processor.CreateThumbnail(context.Background(), request)
				if err != nil {
					b.Fatal(err)
				}
				size = len(result.Content)
			}
			b.ReportMetric(float64(size), "output-bytes")
		})
	}
}
