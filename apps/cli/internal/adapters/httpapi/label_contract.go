package httpapi

import (
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// Keep the REST uint32 versions instead of narrowing through generated int32
// models. Generated SDK methods still own request paths and transport.
type labelRenderBody struct {
	Media    printing.Media      `json:"media"`
	Template labelRenderTemplate `json:"template"`
	Format   string              `json:"format"`
}
type labelRenderTemplate struct {
	ID      string                      `json:"id"`
	Version uint32                      `json:"version"`
	Options ports.LabelTemplateDefaults `json:"options"`
}
type labelEnvelope[T any] struct {
	Schema *string        `json:"$schema,omitempty"`
	Data   T              `json:"data"`
	Meta   generated.Meta `json:"meta"`
}
