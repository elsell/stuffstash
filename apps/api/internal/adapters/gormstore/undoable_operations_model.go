package gormstore

import (
	"encoding/json"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type undoableOperationModel struct {
	ID                string `gorm:"primaryKey;size:26"`
	TenantID          string `gorm:"not null;size:26;index:idx_undoable_operations_scope"`
	Tenant            tenantModel
	InventoryID       string `gorm:"not null;size:26;index:idx_undoable_operations_scope"`
	Inventory         inventoryModel
	PrincipalID       string `gorm:"not null;size:255"`
	Source            string `gorm:"not null;size:64"`
	TargetType        string `gorm:"not null;size:64;index:idx_undoable_operations_target"`
	TargetID          string `gorm:"not null;size:255;index:idx_undoable_operations_target"`
	OriginalAction    string `gorm:"not null;size:96"`
	Status            string `gorm:"not null;size:32;index"`
	CreatedAt         time.Time
	LastAppliedAt     *time.Time
	BeforeAsset       *string `gorm:"type:jsonb"`
	AfterAsset        string  `gorm:"type:jsonb;not null"`
	UndoAuditRecordID *string `gorm:"size:26"`
	RedoAuditRecordID *string `gorm:"size:26"`
}

func (undoableOperationModel) TableName() string {
	return "undoable_operations"
}

type undoableAssetSnapshot struct {
	ID                string         `json:"id"`
	TenantID          string         `json:"tenantId"`
	InventoryID       string         `json:"inventoryId"`
	ParentAssetID     string         `json:"parentAssetId,omitempty"`
	CustomAssetTypeID string         `json:"customAssetTypeId,omitempty"`
	Kind              string         `json:"kind"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	CustomFields      map[string]any `json:"customFields"`
	LifecycleState    string         `json:"lifecycleState"`
}

func newUndoableOperationModel(operation ports.UndoableOperation) (undoableOperationModel, error) {
	afterAsset, err := marshalUndoableAssetSnapshot(operation.AfterAsset)
	if err != nil {
		return undoableOperationModel{}, err
	}
	var beforeAsset *string
	if operation.BeforeAsset != nil {
		encoded, err := marshalUndoableAssetSnapshot(*operation.BeforeAsset)
		if err != nil {
			return undoableOperationModel{}, err
		}
		beforeAsset = &encoded
	}
	var lastAppliedAt *time.Time
	if !operation.LastAppliedAt.IsZero() {
		lastAppliedAt = &operation.LastAppliedAt
	}
	return undoableOperationModel{
		ID:                operation.ID,
		TenantID:          operation.TenantID.String(),
		InventoryID:       operation.InventoryID.String(),
		PrincipalID:       operation.PrincipalID.String(),
		Source:            operation.Source.String(),
		TargetType:        operation.TargetType.String(),
		TargetID:          operation.TargetID,
		OriginalAction:    operation.OriginalAction.String(),
		Status:            string(operation.Status),
		CreatedAt:         operation.CreatedAt,
		LastAppliedAt:     lastAppliedAt,
		BeforeAsset:       beforeAsset,
		AfterAsset:        afterAsset,
		UndoAuditRecordID: stringPtrFromAuditID(operation.UndoAuditRecordID),
		RedoAuditRecordID: stringPtrFromAuditID(operation.RedoAuditRecordID),
	}, nil
}

func (m undoableOperationModel) toPort() (ports.UndoableOperation, bool) {
	afterAsset, ok := unmarshalUndoableAssetSnapshot(m.AfterAsset)
	if !ok {
		return ports.UndoableOperation{}, false
	}
	var beforeAsset *asset.Asset
	if m.BeforeAsset != nil {
		item, ok := unmarshalUndoableAssetSnapshot(*m.BeforeAsset)
		if !ok {
			return ports.UndoableOperation{}, false
		}
		beforeAsset = &item
	}
	lastAppliedAt := time.Time{}
	if m.LastAppliedAt != nil {
		lastAppliedAt = *m.LastAppliedAt
	}
	source, ok := audit.NewSource(m.Source)
	if !ok {
		return ports.UndoableOperation{}, false
	}
	targetType, ok := audit.NewTargetType(m.TargetType)
	if !ok {
		return ports.UndoableOperation{}, false
	}
	action, ok := audit.NewAction(m.OriginalAction)
	if !ok {
		return ports.UndoableOperation{}, false
	}
	undoAuditRecordID := audit.ID("")
	if stringFromPtr(m.UndoAuditRecordID) != "" {
		undoAuditRecordID, ok = audit.NewID(stringFromPtr(m.UndoAuditRecordID))
		if !ok {
			return ports.UndoableOperation{}, false
		}
	}
	redoAuditRecordID := audit.ID("")
	if stringFromPtr(m.RedoAuditRecordID) != "" {
		redoAuditRecordID, ok = audit.NewID(stringFromPtr(m.RedoAuditRecordID))
		if !ok {
			return ports.UndoableOperation{}, false
		}
	}
	return ports.UndoableOperation{
		ID:                m.ID,
		TenantID:          tenant.ID(m.TenantID),
		InventoryID:       inventory.InventoryID(m.InventoryID),
		PrincipalID:       identity.PrincipalID(m.PrincipalID),
		Source:            source,
		TargetType:        targetType,
		TargetID:          m.TargetID,
		OriginalAction:    action,
		Status:            ports.UndoableOperationStatus(m.Status),
		CreatedAt:         m.CreatedAt,
		LastAppliedAt:     lastAppliedAt,
		BeforeAsset:       beforeAsset,
		AfterAsset:        afterAsset,
		UndoAuditRecordID: undoAuditRecordID,
		RedoAuditRecordID: redoAuditRecordID,
	}, true
}

func marshalUndoableAssetSnapshot(item asset.Asset) (string, error) {
	encoded, err := json.Marshal(undoableAssetSnapshot{
		ID:                item.ID.String(),
		TenantID:          item.TenantID.String(),
		InventoryID:       item.InventoryID.String(),
		ParentAssetID:     item.ParentAssetID.String(),
		CustomAssetTypeID: item.CustomAssetTypeID.String(),
		Kind:              item.Kind.String(),
		Title:             item.Title.String(),
		Description:       item.Description.String(),
		CustomFields:      item.CustomFields.Values(),
		LifecycleState:    item.LifecycleState.String(),
	})
	return string(encoded), err
}

func unmarshalUndoableAssetSnapshot(encoded string) (asset.Asset, bool) {
	var snapshot undoableAssetSnapshot
	if err := json.Unmarshal([]byte(encoded), &snapshot); err != nil {
		return asset.Asset{}, false
	}
	id, ok := asset.NewID(snapshot.ID)
	if !ok {
		return asset.Asset{}, false
	}
	kind, ok := asset.NewKind(snapshot.Kind)
	if !ok {
		return asset.Asset{}, false
	}
	title, ok := asset.NewTitle(snapshot.Title)
	if !ok {
		return asset.Asset{}, false
	}
	fields, ok := asset.NewCustomFields(snapshot.CustomFields)
	if !ok {
		return asset.Asset{}, false
	}
	return asset.Asset{
		ID:                id,
		TenantID:          asset.TenantID(snapshot.TenantID),
		InventoryID:       asset.InventoryID(snapshot.InventoryID),
		ParentAssetID:     asset.ID(snapshot.ParentAssetID),
		CustomAssetTypeID: asset.CustomAssetTypeID(snapshot.CustomAssetTypeID),
		Kind:              kind,
		Title:             title,
		Description:       asset.NewDescription(snapshot.Description),
		CustomFields:      fields,
		LifecycleState:    asset.LifecycleState(snapshot.LifecycleState),
	}, true
}
