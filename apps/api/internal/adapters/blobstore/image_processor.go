package blobstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	thumbnailSmallMaxDimension  = 256
	thumbnailMediumMaxDimension = 768
	thumbnailLargeMaxDimension  = 1600
	thumbnailJPEGQuality        = 78
	maxImageInputDimension      = 12000
	maxImageInputPixels         = 50000000
)

type StandardImageProcessor struct{}

func (StandardImageProcessor) CreateThumbnail(_ context.Context, request ports.ImageDerivativeRequest) (ports.ImageDerivative, error) {
	if !request.ContentType.IsImage() || len(request.Content) == 0 {
		return ports.ImageDerivative{}, errors.New("thumbnail source must be an image")
	}
	if err := validateImageBounds(request.Content); err != nil {
		return ports.ImageDerivative{}, err
	}
	source, _, err := image.Decode(bytes.NewReader(request.Content))
	if err != nil {
		return ports.ImageDerivative{}, err
	}
	thumbnail := resizeImage(source, thumbnailMaxDimension(request.Variant))
	output := bytes.Buffer{}
	if err := jpeg.Encode(&output, thumbnail, &jpeg.Options{Quality: thumbnailJPEGQuality}); err != nil {
		return ports.ImageDerivative{}, err
	}
	return ports.ImageDerivative{
		ContentType: media.ContentTypeJPEG,
		Content:     output.Bytes(),
	}, nil
}

func (StandardImageProcessor) PrepareImageForModelUse(_ context.Context, request ports.ModelImageRequest) (ports.ModelImage, error) {
	if !request.ContentType.IsImage() || len(request.Content) == 0 {
		return ports.ModelImage{}, errors.New("model image source must be an image")
	}
	if err := validateImageBounds(request.Content); err != nil {
		return ports.ModelImage{}, err
	}
	source, _, err := image.Decode(bytes.NewReader(request.Content))
	if err != nil {
		return ports.ModelImage{}, err
	}
	hashBytes := sha256.Sum256(request.Content)
	hash, ok := media.NewSHA256(hex.EncodeToString(hashBytes[:]))
	if !ok {
		return ports.ModelImage{}, errors.New("model image hash invalid")
	}
	return ports.ModelImage{
		ContentType: request.ContentType,
		Content:     append([]byte(nil), request.Content...),
		SizeBytes:   int64(len(request.Content)),
		SHA256:      hash,
		Width:       source.Bounds().Dx(),
		Height:      source.Bounds().Dy(),
	}, nil
}

func validateImageBounds(content []byte) error {
	config, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		return err
	}
	if config.Width <= 0 || config.Height <= 0 {
		return errors.New("image dimensions invalid")
	}
	if config.Width > maxImageInputDimension || config.Height > maxImageInputDimension {
		return errors.New("image dimensions too large")
	}
	if config.Width > maxImageInputPixels/config.Height {
		return errors.New("image pixel count too large")
	}
	return nil
}

func thumbnailMaxDimension(variant media.ThumbnailVariant) int {
	switch variant {
	case media.ThumbnailVariantMedium:
		return thumbnailMediumMaxDimension
	case media.ThumbnailVariantLarge:
		return thumbnailLargeMaxDimension
	default:
		return thumbnailSmallMaxDimension
	}
}

func resizeImage(source image.Image, maxDimension int) image.Image {
	bounds := source.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	if width <= maxDimension && height <= maxDimension {
		return copyImage(source)
	}

	scale := float64(maxDimension) / float64(width)
	if height > width {
		scale = float64(maxDimension) / float64(height)
	}
	targetWidth := max(1, int(float64(width)*scale))
	targetHeight := max(1, int(float64(height)*scale))
	target := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(target, target.Bounds(), source, bounds, draw.Over, nil)
	return target
}

func copyImage(source image.Image) image.Image {
	bounds := source.Bounds()
	target := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := range bounds.Dy() {
		for x := range bounds.Dx() {
			target.Set(x, y, source.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return target
}

func init() {
	image.RegisterFormat("jpeg", "\xff\xd8", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
}
