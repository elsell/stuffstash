package printing

type TemplateID string

const (
	TemplateQRTitle TemplateID = "qr-title"
	TemplateQROnly  TemplateID = "qr-only"
)

type TemplateOptions struct {
	ShowReference bool `json:"show_reference"`
}
type TemplateSelection struct {
	ID      TemplateID      `json:"id"`
	Version uint32          `json:"version"`
	Options TemplateOptions `json:"options"`
}
type TemplateDescriptor struct {
	ID                    TemplateID      `json:"id"`
	Version               uint32          `json:"version"`
	Name                  string          `json:"name"`
	Purpose               string          `json:"purpose"`
	Options               []string        `json:"options"`
	Defaults              TemplateOptions `json:"defaults"`
	MinimumQRModulePixels int             `json:"minimum_qr_module_pixels"`
	Font                  string          `json:"font"`
	GlyphCoverage         string          `json:"glyph_coverage"`
}
