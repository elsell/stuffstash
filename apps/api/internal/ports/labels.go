package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"time"
)

type LabelRepository interface {
	LabelInstance(context.Context) (printing.InstanceID, bool, error)
	BootstrapLabelInstance(context.Context, printing.InstanceID) (printing.InstanceID, error)
	LabelForAsset(context.Context, tenant.ID, inventory.InventoryID, asset.ID) (printing.Label, bool, error)
	// LookupLabel is the specified opaque-ID resolver exception. Callers must
	// authorize its returned inventory before exposing any resource information.
	LookupLabel(context.Context, printing.InstanceID, printing.LabelID) (printing.Label, bool, error)
	ProvisionLabel(context.Context, printing.Label, audit.Record) (printing.Label, bool, error)
}
type LabelRenderRepository interface {
	SaveLabelRender(context.Context, printing.LabelRender, audit.Record) error
	LabelRenderByID(context.Context, tenant.ID, inventory.InventoryID, printing.RenderID) (printing.LabelRender, bool, error)
	PurgeExpiredLabelRenders(context.Context, time.Time) error
}
