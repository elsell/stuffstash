package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type AuthInput struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
}
type ScopeInput struct {
	AuthInput
	TenantID    string `path:"tenantId"`
	InventoryID string `path:"inventoryId"`
}
type AssetInput struct {
	ScopeInput
	AssetID string `path:"assetId"`
}
type ResolveInput struct {
	AuthInput
	InstanceID string `path:"instanceId"`
	LabelID    string `path:"labelId"`
}
type InstanceResponse struct {
	ProtocolVersion int    `json:"protocolVersion"`
	InstanceID      string `json:"instanceId"`
}
type InstanceOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[InstanceResponse]
}
type LabelResponse struct {
	LabelID        string `json:"labelId"`
	InstanceID     string `json:"instanceId"`
	TenantID       string `json:"tenantId"`
	InventoryID    string `json:"inventoryId"`
	AssetID        string `json:"assetId"`
	URL            string `json:"url"`
	LifecycleState string `json:"lifecycleState"`
}
type LabelOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[LabelResponse]
}
type Margins struct {
	Left   int `json:"left"`
	Right  int `json:"right"`
	Top    int `json:"top"`
	Bottom int `json:"bottom"`
}
type Media struct {
	PresetID           string  `json:"preset_id,omitempty"`
	Version            uint32  `json:"version,omitempty"`
	WidthMicrometers   int     `json:"width_micrometers"`
	HeightMicrometers  int     `json:"height_micrometers"`
	MarginsMicrometers Margins `json:"margins_micrometers"`
	ResolutionDPI      int     `json:"resolution_dpi"`
	RasterWidth        int     `json:"raster_width"`
	RasterHeight       int     `json:"raster_height"`
	Orientation        string  `json:"orientation"`
	ColorMode          string  `json:"color_mode"`
	CutPolicy          string  `json:"cut_policy"`
	DisplayRotation    int     `json:"display_rotation"`
}
type TemplateOptions struct {
	ShowReference bool `json:"show_reference"`
}
type TemplateSelection struct {
	ID      string          `json:"id"`
	Version uint32          `json:"version"`
	Options TemplateOptions `json:"options"`
}
type RenderInput struct {
	AssetInput
	Body struct {
		Media    Media             `json:"media"`
		Template TemplateSelection `json:"template"`
		Format   string            `json:"format" enum:"png,pdf"`
	}
}
type RenderResponse struct {
	ID                   string `json:"id"`
	SelectionFingerprint string `json:"selectionFingerprint"`
	MediaFingerprint     string `json:"mediaFingerprint"`
	ContentType          string `json:"contentType"`
	SHA256               string `json:"sha256"`
	ContentPath          string `json:"contentPath"`
	ExpiresAt            string `json:"expiresAt"`
	WidthPixels          int    `json:"widthPixels"`
	HeightPixels         int    `json:"heightPixels"`
	DisplayRotation      int    `json:"displayRotation"`
}
type RenderOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[RenderResponse]
}
type ContentInput struct {
	ScopeInput
	RenderID string `path:"renderId"`
}
type ContentOutput struct {
	CacheControl       string `header:"Cache-Control"`
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
	Body               []byte
}
type TemplateResponse struct {
	ID                    string          `json:"id"`
	Version               uint32          `json:"version"`
	Name                  string          `json:"name"`
	Purpose               string          `json:"purpose"`
	Options               []string        `json:"options"`
	Defaults              TemplateOptions `json:"defaults"`
	Font                  string          `json:"font"`
	GlyphCoverage         string          `json:"glyphCoverage"`
	MinimumQRModulePixels int             `json:"minimumQRModulePixels"`
}
type TemplatesOutput struct {
	Body shared.SuccessEnvelope[[]TemplateResponse]
}
