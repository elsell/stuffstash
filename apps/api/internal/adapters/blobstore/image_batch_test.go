package blobstore

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestBatchThumbnailsPublishInSizeOrderWithEquivalentOutput(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 321, 241))
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, source); err != nil {
		t.Fatal(err)
	}
	request := ports.ImageDerivativesRequest{ContentType: media.ContentTypePNG, Content: encoded.Bytes(), Variants: []media.ThumbnailVariant{media.ThumbnailVariantLarge, media.ThumbnailVariantSmall, media.ThumbnailVariantMedium}}
	var order []media.ThumbnailVariant
	err := (StandardImageProcessor{}).CreateThumbnails(context.Background(), request, func(variant media.ThumbnailVariant, derivative ports.ImageDerivative) error {
		order = append(order, variant)
		single, err := (StandardImageProcessor{}).CreateThumbnail(context.Background(), ports.ImageDerivativeRequest{ContentType: request.ContentType, Content: request.Content, Variant: variant})
		if err != nil {
			return err
		}
		if !bytes.Equal(single.Content, derivative.Content) || single.ContentType != derivative.ContentType {
			t.Fatal("batch output differs from foreground output")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 3 || order[0] != media.ThumbnailVariantSmall || order[1] != media.ThumbnailVariantMedium || order[2] != media.ThumbnailVariantLarge {
		t.Fatal("variants were not published in readiness order")
	}
}

func TestBatchStopsOnPublicationFailureAndCancellation(t *testing.T) {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	for _, cancellation := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		count := 0
		failure := errors.New("controlled persistence failure")
		err := (StandardImageProcessor{}).CreateThumbnails(ctx, ports.ImageDerivativesRequest{ContentType: media.ContentTypePNG, Content: encoded.Bytes(), Variants: []media.ThumbnailVariant{media.ThumbnailVariantSmall, media.ThumbnailVariantMedium}}, func(media.ThumbnailVariant, ports.ImageDerivative) error {
			count++
			if cancellation {
				cancel()
				return nil
			}
			return failure
		})
		cancel()
		expected := failure
		if cancellation {
			expected = context.Canceled
		}
		if !errors.Is(err, expected) || count != 1 {
			t.Fatalf("batch continued after failure/cancellation: %v, %d", err, count)
		}
	}
}

func TestBatchRejectsInvalidVariantSetBeforePublication(t *testing.T) {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	for _, variants := range [][]media.ThumbnailVariant{nil, {media.ThumbnailVariantSmall, media.ThumbnailVariantSmall}, {media.ThumbnailVariantSmall, "invalid"}} {
		count := 0
		err := (StandardImageProcessor{}).CreateThumbnails(context.Background(), ports.ImageDerivativesRequest{ContentType: media.ContentTypePNG, Content: encoded.Bytes(), Variants: variants}, func(media.ThumbnailVariant, ports.ImageDerivative) error { count++; return nil })
		if err == nil || count != 0 {
			t.Fatal("invalid batch was published")
		}
	}
}

func TestSingleThumbnailRetainsEmptyVariantDefault(t *testing.T) {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 321, 241))); err != nil {
		t.Fatal(err)
	}
	result, err := (StandardImageProcessor{}).CreateThumbnail(context.Background(), ports.ImageDerivativeRequest{ContentType: media.ContentTypePNG, Content: encoded.Bytes()})
	if err != nil {
		t.Fatal(err)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(result.Content))
	if err != nil || config.Width != 256 {
		t.Fatalf("empty variant lost small default: %v %#v", err, config)
	}
}
