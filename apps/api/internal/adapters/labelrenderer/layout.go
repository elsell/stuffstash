package labelrenderer

import (
	"errors"
	"image"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func composeTitle(dst *image.Gray, c printing.ContentSnapshot, opts printing.TemplateOptions, typeface *opentype.Font) error {
	bounds := dst.Bounds()
	gap := max(8, min(bounds.Dx(), bounds.Dy())/24)
	area := bounds.Inset(gap)
	side := min(area.Dy(), area.Dx()/2)
	if side < 1 {
		return errors.New("printable area is too small")
	}
	qrArea := image.Rect(area.Min.X, area.Min.Y, area.Min.X+side, area.Min.Y+side)
	if err := paintQR(dst, qrArea, c.QRURL); err != nil {
		return err
	}
	textArea := image.Rect(qrArea.Max.X+gap, area.Min.Y, area.Max.X, area.Max.Y)
	fontSize := max(12, min(42, area.Dy()/6))
	referenceHeight := 0
	if opts.ShowReference && c.Reference != "" {
		referenceHeight = fontSize + gap
	}
	if err := drawText(dst, image.Rect(textArea.Min.X, textArea.Min.Y, textArea.Max.X, textArea.Max.Y-referenceHeight), c.Title, fontSize, typeface, false); err != nil {
		return err
	}
	if referenceHeight > 0 {
		return drawText(dst, image.Rect(textArea.Min.X, textArea.Max.Y-referenceHeight+gap, textArea.Max.X, textArea.Max.Y), c.Reference, max(10, fontSize*2/3), typeface, false)
	}
	return nil
}
func composeOnly(dst *image.Gray, c printing.ContentSnapshot, opts printing.TemplateOptions, typeface *opentype.Font) error {
	gap := max(8, min(dst.Bounds().Dx(), dst.Bounds().Dy())/24)
	area := dst.Bounds().Inset(gap)
	referenceHeight := 0
	size := max(12, min(28, area.Dy()/8))
	if opts.ShowReference && c.Reference != "" {
		referenceHeight = size*3/2 + gap
	}
	if err := paintQR(dst, image.Rect(area.Min.X, area.Min.Y, area.Max.X, area.Max.Y-referenceHeight), c.QRURL); err != nil {
		return err
	}
	if referenceHeight > 0 {
		return drawText(dst, image.Rect(area.Min.X, area.Max.Y-referenceHeight+gap, area.Max.X, area.Max.Y), c.Reference, size, typeface, true)
	}
	return nil
}

func drawText(dst *image.Gray, area image.Rectangle, text string, size int, typeface *opentype.Font, center bool) error {
	if text == "" {
		return nil
	}
	face, err := opentype.NewFace(typeface, &opentype.FaceOptions{Size: float64(size), DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return err
	}
	defer face.Close()
	lineHeight := face.Metrics().Height.Ceil()
	count := area.Dy() / lineHeight
	if count < 1 || area.Dx() < font.MeasureString(face, "…").Ceil() {
		return errors.New("label has insufficient printable space for text")
	}
	lines := wrapText(face, text, area.Dx(), count)
	drawer := font.Drawer{Dst: dst, Src: image.Black, Face: face}
	baseline := area.Min.Y + (area.Dy()-lineHeight*len(lines))/2 + face.Metrics().Ascent.Ceil()
	for _, line := range lines {
		x := area.Min.X
		if center {
			x += (area.Dx() - font.MeasureString(face, line).Ceil()) / 2
		}
		drawer.Dot = fixed.P(x, baseline)
		drawer.DrawString(line)
		baseline += lineHeight
	}
	return nil
}
func wrapText(face font.Face, text string, width, maxLines int) []string {
	remaining := []rune(text)
	lines := make([]string, 0, maxLines)
	for len(remaining) > 0 && len(lines) < maxLines {
		fit := 0
		lastSpace := -1
		for i, r := range remaining {
			if font.MeasureString(face, string(remaining[:i+1])).Ceil() > width {
				break
			}
			fit = i + 1
			if r == ' ' {
				lastSpace = i
			}
		}
		if fit == len(remaining) {
			lines = append(lines, string(remaining))
			break
		}
		if len(lines) == maxLines-1 {
			for fit > 0 && font.MeasureString(face, string(remaining[:fit])+"…").Ceil() > width {
				fit--
			}
			lines = append(lines, strings.TrimSpace(string(remaining[:fit]))+"…")
			break
		}
		if lastSpace > 0 {
			fit = lastSpace
		}
		if fit == 0 {
			lines = append(lines, "…")
			break
		}
		lines = append(lines, strings.TrimSpace(string(remaining[:fit])))
		remaining = []rune(strings.TrimSpace(string(remaining[fit:])))
	}
	return lines
}
