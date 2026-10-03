package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/labels"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
)

type LabelTemplate struct {
	ID            string `json:"id"`
	Version       uint32 `json:"version"`
	Name          string `json:"name"`
	Purpose       string `json:"purpose"`
	ShowReference bool   `json:"showReference"`
}
type ResolvedLabel struct {
	TenantID    string `json:"tenantId"`
	InventoryID string `json:"inventoryId"`
	AssetID     string `json:"assetId"`
	Lifecycle   string `json:"lifecycleState"`
}
type LabelRenderSelection struct {
	TemplateID      string
	TemplateVersion uint32
	ShowReference   bool
	Media           printing.Media
	Format          string
}
type LabelArtifact struct {
	Content        []byte
	Format, SHA256 string
}
type LabelFileResult struct {
	Path   string `json:"path"`
	Format string `json:"format"`
	SHA256 string `json:"sha256"`
}
type LabelsAPI interface {
	PrintDefaults(context.Context, Scope) (InventoryPrintDefaults, error)
	LabelTemplates(context.Context, Scope) (Result[[]LabelTemplate], error)
	LabelMedia(context.Context, Scope, string) ([]printing.Media, error)
	ResolveLabel(context.Context, labels.Reference) (Result[ResolvedLabel], error)
	RenderLabel(context.Context, Scope, string, LabelRenderSelection) (LabelArtifact, error)
}
type LabelFiles interface {
	Publish(context.Context, string, []byte) error
}
