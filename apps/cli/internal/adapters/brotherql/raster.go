package brotherql

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"image/png"
)

// encode preserves the image composition. The horizontal reversal is Brother's
// physical pin order, not a layout transform: pin zero is at the right edge.
func encode(label printing.Label) ([]byte, error) {
	if label.AttemptID == "" || label.Copy < 1 || label.Media != Media() || len(label.PNG) > 4<<20 {
		return nil, errors.New("unsupported label profile or input")
	}
	config, err := png.DecodeConfig(bytes.NewReader(label.PNG))
	if err != nil || config.Width != rasterWidth || config.Height != rasterHeight {
		return nil, errors.New("PNG dimensions do not match the registered media")
	}
	image, err := png.Decode(bytes.NewReader(label.PNG))
	if err != nil {
		return nil, errors.New("invalid PNG")
	}
	data := make([]byte, 400, 400+64+rasterHeight*(3+rasterBytes))
	data = append(data, 0x1b, 0x40, 0x1b, 0x69, 0x61, 1, 0x1b, 0x69, 0x21, 0)
	// Media type/width/length valid, quality priority, printer recovery on.
	data = append(data, 0x1b, 0x69, 0x7a, 0xce, 0x0b, 29, 90)
	data = binary.LittleEndian.AppendUint32(data, rasterHeight)
	data = append(data, 0, 0, 0x1b, 0x69, 0x4d, 0x40, 0x1b, 0x69, 0x41, 1, 0x1b, 0x69, 0x4b, 8, 0x1b, 0x69, 0x64, 0, 0)
	// QL-800 supports neither compression nor the zero-raster shorthand.
	for y := 0; y < rasterHeight; y++ {
		row := [rasterBytes]byte{}
		for x := 0; x < rasterWidth; x++ {
			r, g, b, a := image.At(x, y).RGBA()
			if a != 65535 || r != g || g != b || (r != 0 && r != 65535) {
				return nil, errors.New("PNG must be opaque monochrome; server rendering is required")
			}
			if r == 0 {
				pin := rightHeadOffset + rasterWidth - 1 - x
				row[pin/8] |= 1 << uint(7-pin%8)
			}
		}
		data = append(data, 0x67, 0, rasterBytes)
		data = append(data, row[:]...)
	}
	return append(data, 0x1a), nil
}
