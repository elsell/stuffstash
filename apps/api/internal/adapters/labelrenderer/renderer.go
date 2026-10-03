// Package labelrenderer composes immutable labels. It never contacts a printer.
package labelrenderer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/boombuler/barcode/qr"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/text/unicode/norm"
)

const minimumModulePixels = 3
const quietModules = 4
const templateVersion = 1

type Limits struct{ MaxPixels, MaxURLBytes, MaxTitleRunes, MaxReferenceRunes int }

// DefaultLimits is for offline tooling. API bootstrap must pass validated limits.
func DefaultLimits() Limits {
	return Limits{MaxPixels: 4000000, MaxURLBytes: 2048, MaxTitleRunes: 2048, MaxReferenceRunes: 128}
}

type registeredTemplate struct {
	descriptor printing.TemplateDescriptor
	compose    func(*image.Gray, printing.ContentSnapshot, printing.TemplateOptions, *opentype.Font) error
}
type Renderer struct {
	limits    Limits
	font      *opentype.Font
	templates []registeredTemplate
}

func New(limits Limits) (*Renderer, error) {
	if limits.MaxPixels <= 0 || limits.MaxURLBytes <= 0 || limits.MaxTitleRunes <= 0 || limits.MaxReferenceRunes <= 0 {
		return nil, errors.New("label rendering limits must be positive")
	}
	face, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse label font: %w", err)
	}
	descriptors := []registeredTemplate{
		{descriptor: descriptor(printing.TemplateQRTitle, "QR and title", "QR beside the asset title"), compose: composeTitle},
		{descriptor: descriptor(printing.TemplateQROnly, "QR", "QR with an optional human-readable reference"), compose: composeOnly},
	}
	return &Renderer{limits: limits, font: face, templates: descriptors}, nil
}
func descriptor(id printing.TemplateID, name, purpose string) printing.TemplateDescriptor {
	return printing.TemplateDescriptor{ID: id, Version: templateVersion, Name: name, Purpose: purpose, Options: []string{"show_reference"}, Defaults: printing.TemplateOptions{ShowReference: true}, MinimumQRModulePixels: minimumModulePixels, Font: "Go Regular (golang.org/x/image v0.41.0)", GlyphCoverage: "Latin, Greek, Cyrillic; unsupported glyphs and shaping scripts are rejected"}
}
func (r *Renderer) Templates() []printing.TemplateDescriptor {
	result := make([]printing.TemplateDescriptor, len(r.templates))
	for i, t := range r.templates {
		result[i] = t.descriptor
		result[i].Options = append([]string(nil), t.descriptor.Options...)
	}
	return result
}
func (r *Renderer) Render(ctx context.Context, req printing.RenderRequest) (printing.RenderedLabel, error) {
	if err := ctx.Err(); err != nil {
		return printing.RenderedLabel{}, err
	}
	if err := req.Media.Validate(r.limits.MaxPixels); err != nil {
		return printing.RenderedLabel{}, err
	}
	if req.Format != printing.FormatPNG && req.Format != printing.FormatPDF {
		return printing.RenderedLabel{}, errors.New("unsupported label format")
	}
	var selected *registeredTemplate
	for i := range r.templates {
		candidate := &r.templates[i]
		if candidate.descriptor.ID == req.Template.ID && candidate.descriptor.Version == req.Template.Version {
			selected = candidate
			break
		}
	}
	if selected == nil {
		return printing.RenderedLabel{}, errors.New("unknown or disabled label template version")
	}
	if err := r.validateContent(req.Content, req.Template); err != nil {
		return printing.RenderedLabel{}, err
	}
	req.Content.Title = norm.NFC.String(strings.Join(strings.Fields(req.Content.Title), " "))
	req.Content.Reference = norm.NFC.String(strings.Join(strings.Fields(req.Content.Reference), " "))
	w, h := req.Media.RasterWidth, req.Media.RasterHeight
	if req.Media.DisplayRotation == 90 || req.Media.DisplayRotation == 270 {
		w, h = h, w
	}
	canvas := image.NewGray(image.Rect(0, 0, w, h))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	if err := selected.compose(canvas, req.Content, req.Template.Options, r.font); err != nil {
		return printing.RenderedLabel{}, err
	}
	for i, pixel := range canvas.Pix {
		if pixel < 128 {
			canvas.Pix[i] = 0
		} else {
			canvas.Pix[i] = 255
		}
	}
	raster := rotate(canvas, (360-req.Media.DisplayRotation)%360)
	var encoded bytes.Buffer
	contentType := "image/png"
	if req.Format == printing.FormatPDF {
		contentType = "application/pdf"
		if err := encodePDF(&encoded, raster, req.Media); err != nil {
			return printing.RenderedLabel{}, err
		}
	} else if err := png.Encode(&encoded, raster); err != nil {
		return printing.RenderedLabel{}, err
	}
	if err := ctx.Err(); err != nil {
		return printing.RenderedLabel{}, err
	}
	content := encoded.Bytes()
	sum := sha256.Sum256(content)
	return printing.RenderedLabel{Content: content, ContentType: contentType, WidthPixels: raster.Bounds().Dx(), HeightPixels: raster.Bounds().Dy(), SHA256: hex.EncodeToString(sum[:]), DisplayRotation: req.Media.DisplayRotation}, nil
}
func (r *Renderer) validateContent(c printing.ContentSnapshot, t printing.TemplateSelection) error {
	if len(c.QRURL) > r.limits.MaxURLBytes || utf8.RuneCountInString(c.Title) > r.limits.MaxTitleRunes || utf8.RuneCountInString(c.Reference) > r.limits.MaxReferenceRunes {
		return errors.New("label content exceeds configured limits")
	}
	if !utf8.ValidString(c.QRURL) || !utf8.ValidString(c.Title) || !utf8.ValidString(c.Reference) {
		return errors.New("label content is not valid UTF-8")
	}
	u, err := url.Parse(c.QRURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("label QR must be an absolute HTTPS URL without credentials, query, or fragment")
	}
	text := c.Reference
	if !t.Options.ShowReference {
		text = ""
	}
	if t.ID == printing.TemplateQRTitle {
		text += "\n" + c.Title
	}
	normalized := norm.NFC.String(text)
	var buffer sfnt.Buffer
	for _, ch := range normalized {
		if unicode.IsSpace(ch) {
			continue
		}
		if unicode.IsControl(ch) || unicode.Is(unicode.Cf, ch) || unicode.Is(unicode.Mn, ch) || unicode.Is(unicode.Mc, ch) {
			return errors.New("label text requires unsupported shaping or contains control characters")
		}
		glyph, err := r.font.GlyphIndex(&buffer, ch)
		if err != nil || glyph == 0 {
			return fmt.Errorf("label font does not support character U+%04X; choose QR-only or supported text", ch)
		}
	}
	return nil
}

func paintQR(dst *image.Gray, area image.Rectangle, value string) error {
	matrix, err := qr.Encode(value, qr.M, qr.Unicode)
	if err != nil {
		return fmt.Errorf("encode label QR: %w", err)
	}
	modules := matrix.Bounds().Dx()
	total := modules + 2*quietModules
	scale := min(area.Dx(), area.Dy()) / total
	if scale < minimumModulePixels {
		return errors.New("QR is too dense for the printable area; use a shorter label address, a simpler template, or larger media")
	}
	x := area.Min.X + (area.Dx()-total*scale)/2 + quietModules*scale
	y := area.Min.Y + (area.Dy()-total*scale)/2 + quietModules*scale
	for row := 0; row < modules; row++ {
		for col := 0; col < modules; col++ {
			gray := color.GrayModel.Convert(matrix.At(col, row)).(color.Gray)
			if gray.Y < 128 {
				draw.Draw(dst, image.Rect(x+col*scale, y+row*scale, x+(col+1)*scale, y+(row+1)*scale), image.Black, image.Point{}, draw.Src)
			}
		}
	}
	return nil
}
func rotate(src *image.Gray, degrees int) *image.Gray {
	if degrees == 0 {
		return src
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	nw, nh := w, h
	if degrees == 90 || degrees == 270 {
		nw, nh = h, w
	}
	dst := image.NewGray(image.Rect(0, 0, nw, nh))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			switch degrees {
			case 90:
				dst.SetGray(h-1-y, x, src.GrayAt(x, y))
			case 180:
				dst.SetGray(w-1-x, h-1-y, src.GrayAt(x, y))
			case 270:
				dst.SetGray(y, w-1-x, src.GrayAt(x, y))
			}
		}
	}
	return dst
}

// DisplayPNG applies only the declared integer rotation; it never resamples.
func (r *Renderer) DisplayPNG(label printing.RenderedLabel) ([]byte, error) {
	if label.ContentType != "image/png" {
		return nil, errors.New("display transform requires PNG")
	}
	decoded, err := png.Decode(bytes.NewReader(label.Content))
	if err != nil {
		return nil, err
	}
	gray := image.NewGray(decoded.Bounds())
	draw.Draw(gray, gray.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	var output bytes.Buffer
	if err := png.Encode(&output, rotate(gray, label.DisplayRotation)); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
