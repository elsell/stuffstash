package httpapi

import (
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// The REST DTO uses unsigned counters and versions; generated response models
// narrow them to signed values. Keep SDK requests, decode into owned exact types.
type printerEnvelope[T any] struct {
	Schema *string        `json:"$schema,omitempty"`
	Data   T              `json:"data"`
	Meta   generated.Meta `json:"meta"`
}

func completePrinter(p ports.RegisteredPrinter) ports.RegisteredPrinter {
	p.MediaName = p.Media.Name
	p.MediaPreset = p.Media.PresetID
	return p
}
