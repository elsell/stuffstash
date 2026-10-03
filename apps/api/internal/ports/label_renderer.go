package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

type LabelRenderer interface {
	Render(context.Context, printing.RenderRequest) (printing.RenderedLabel, error)
}

type LabelTemplateCatalog interface {
	Templates() []printing.TemplateDescriptor
}

// LabelDisplayTransformer only changes reading orientation, never composition.
type LabelDisplayTransformer interface {
	DisplayPNG(printing.RenderedLabel) ([]byte, error)
}
