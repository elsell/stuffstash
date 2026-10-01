package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type PreparedImportAsset struct {
	Asset                 asset.Asset
	AuditRecord           audit.Record
	PromotedParent        *asset.Asset
	ParentPromotionRecord *audit.Record
	UndoableOperation     UndoableOperation
}

type PreparedImportAttachment struct {
	ThumbnailJob *media.ThumbnailJob
	Attachment   media.Attachment
	AuditRecord  audit.Record
	StorageKey   media.StorageKey
	ContentType  media.ContentType
}

// ImportTargets preserves target-domain validation while the import unit of work
// atomically persists prepared records together with import links and history.
type ImportTargets interface {
	EnsureActiveInventoryAccess(context.Context, identity.Principal, tenant.ID, inventory.InventoryID, InventoryPermission) error
	CreateField(context.Context, ImportJobCommand, importplan.FieldDefinition) error
	CreateTag(context.Context, ImportJobCommand, importplan.TagDefinition) (assettag.Tag, error)
	PrepareAsset(context.Context, ImportJobCommand, importplan.Asset, string) (PreparedImportAsset, error)
	PrepareAttachment(context.Context, ImportJobCommand, asset.ID, importplan.Attachment) (PreparedImportAttachment, error)
	SetAssetTags(context.Context, ImportJobCommand, asset.ID, []string) error
	RecordAssetCreated(context.Context, asset.Asset, identity.PrincipalID)
	RecordAttachmentCreated(context.Context, ImportJobCommand, asset.ID, media.Attachment)
	DeleteAsset(context.Context, ImportJobCommand, asset.ID) error
	DeleteAttachment(context.Context, ImportJobCommand, asset.ID, media.ID) error
}
