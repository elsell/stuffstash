package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/labels"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
)

type LabelTemplateDefaults struct {
	ShowReference bool `json:"show_reference"`
}
type LabelTemplate struct {
	Defaults              LabelTemplateDefaults `json:"defaults"`
	Font                  string                `json:"font"`
	GlyphCoverage         string                `json:"glyphCoverage"`
	MinimumQRModulePixels int64                 `json:"minimumQRModulePixels"`
	Options               []string              `json:"options"`
	ID                    string                `json:"id"`
	Version               uint32                `json:"version"`
	Name                  string                `json:"name"`
	Purpose               string                `json:"purpose"`
	ShowReference         bool                  `json:"showReference"`
}
type ResolvedLabel struct {
	InstanceID  string `json:"instanceId"`
	LabelID     string `json:"labelId"`
	URL         string `json:"url"`
	TenantID    string `json:"tenantId"`
	InventoryID string `json:"inventoryId"`
	AssetID     string `json:"assetId"`
	Lifecycle   string `json:"lifecycleState"`
}
type LabelRenderSelection struct {
	RequestBody     []byte
	TemplateID      string
	TemplateVersion uint32
	ShowReference   bool
	Media           printing.Media
	Format          string
}
type LabelRenderMetadata struct {
	ID                   string `json:"id"`
	SelectionFingerprint string `json:"selectionFingerprint"`
	MediaFingerprint     string `json:"mediaFingerprint"`
	ContentType          string `json:"contentType"`
	SHA256               string `json:"sha256"`
	ContentPath          string `json:"contentPath"`
	ExpiresAt            string `json:"expiresAt"`
	WidthPixels          int64  `json:"widthPixels"`
	HeightPixels         int64  `json:"heightPixels"`
	DisplayRotation      int64  `json:"displayRotation"`
}
type LabelArtifact struct {
	Render         Result[LabelRenderMetadata]
	Content        []byte
	Format, SHA256 string
}
type LabelFileResult struct {
	Render Result[LabelRenderMetadata] `json:"render"`
	Path   string                      `json:"path"`
	Format string                      `json:"format"`
	SHA256 string                      `json:"sha256"`
}
type LabelsAPI interface {
	AssetLabel(context.Context, Scope, string) (Result[ResolvedLabel], error)
	AssignLabel(context.Context, Scope, string) (Result[ResolvedLabel], error)
	PrintDefaults(context.Context, Scope) (InventoryPrintDefaults, error)
	LabelTemplates(context.Context, Scope) (Result[[]LabelTemplate], error)
	LabelMedia(context.Context, Scope, string) ([]printing.Media, error)
	ResolveLabel(context.Context, labels.Reference) (Result[ResolvedLabel], error)
	RenderLabel(context.Context, Scope, string, LabelRenderSelection) (LabelArtifact, error)
}
type LabelFiles interface {
	Publish(context.Context, string, []byte) error
}
