package gormstore

import (
	"encoding/json"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
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
	AfterAsset        *string `gorm:"type:jsonb"`
	BeforeCheckout    *string `gorm:"type:jsonb"`
	AfterCheckout     *string `gorm:"type:jsonb"`
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
	TagIDs            []string       `json:"tagIds,omitempty"`
	ReplacesTags      bool           `json:"replacesTags,omitempty"`
}

type undoableCheckoutSnapshot struct {
	ID                    string `json:"id"`
	TenantID              string `json:"tenantId"`
	InventoryID           string `json:"inventoryId"`
	AssetID               string `json:"assetId"`
	State                 string `json:"state"`
	CheckedOutAt          string `json:"checkedOutAt"`
	CheckedOutByPrincipal string `json:"checkedOutByPrincipal"`
	CheckoutDetails       string `json:"checkoutDetails,omitempty"`
	ReturnedAt            string `json:"returnedAt,omitempty"`
	ReturnedByPrincipal   string `json:"returnedByPrincipal,omitempty"`
	ReturnDetails         string `json:"returnDetails,omitempty"`
	CreatedAt             string `json:"createdAt"`
	UpdatedAt             string `json:"updatedAt"`
}

func newUndoableOperationModel(operation ports.UndoableOperation) (undoableOperationModel, error) {
	var afterAsset *string
	var beforeAsset *string
	if operation.AfterCheckout == nil {
		encoded, err := marshalUndoableAssetSnapshot(operation.AfterAsset, operation.AfterTagIDs, operation.ReplacesTags)
		if err != nil {
			return undoableOperationModel{}, err
		}
		afterAsset = &encoded
		if operation.BeforeAsset != nil {
			encoded, err := marshalUndoableAssetSnapshot(*operation.BeforeAsset, operation.BeforeTagIDs, operation.ReplacesTags)
			if err != nil {
				return undoableOperationModel{}, err
			}
			beforeAsset = &encoded
		}
	}
	var afterCheckout *string
	var beforeCheckout *string
	if operation.AfterCheckout != nil {
		encoded, err := marshalUndoableCheckoutSnapshot(*operation.AfterCheckout)
		if err != nil {
			return undoableOperationModel{}, err
		}
		afterCheckout = &encoded
		if operation.BeforeCheckout != nil {
			encoded, err := marshalUndoableCheckoutSnapshot(*operation.BeforeCheckout)
			if err != nil {
				return undoableOperationModel{}, err
			}
			beforeCheckout = &encoded
		}
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
		BeforeCheckout:    beforeCheckout,
		AfterCheckout:     afterCheckout,
		UndoAuditRecordID: stringPtrFromAuditID(operation.UndoAuditRecordID),
		RedoAuditRecordID: stringPtrFromAuditID(operation.RedoAuditRecordID),
	}, nil
}

func (m undoableOperationModel) toPort() (ports.UndoableOperation, bool) {
	afterAsset := asset.Asset{}
	var afterTagIDs []assettag.ID
	replacesTags := false
	if m.AfterAsset != nil {
		var ok bool
		afterAsset, afterTagIDs, replacesTags, ok = unmarshalUndoableAssetSnapshot(*m.AfterAsset)
		if !ok {
			return ports.UndoableOperation{}, false
		}
	}
	var beforeAsset *asset.Asset
	var beforeTagIDs []assettag.ID
	if m.BeforeAsset != nil {
		item, tagIDs, beforeReplacesTags, ok := unmarshalUndoableAssetSnapshot(*m.BeforeAsset)
		if !ok {
			return ports.UndoableOperation{}, false
		}
		beforeAsset = &item
		beforeTagIDs = tagIDs
		replacesTags = replacesTags || beforeReplacesTags
	}
	var afterCheckout *asset.Checkout
	if m.AfterCheckout != nil {
		checkout, ok := unmarshalUndoableCheckoutSnapshot(*m.AfterCheckout)
		if !ok {
			return ports.UndoableOperation{}, false
		}
		afterCheckout = &checkout
	}
	var beforeCheckout *asset.Checkout
	if m.BeforeCheckout != nil {
		checkout, ok := unmarshalUndoableCheckoutSnapshot(*m.BeforeCheckout)
		if !ok {
			return ports.UndoableOperation{}, false
		}
		beforeCheckout = &checkout
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
		BeforeTagIDs:      beforeTagIDs,
		AfterTagIDs:       afterTagIDs,
		ReplacesTags:      replacesTags,
		BeforeCheckout:    beforeCheckout,
		AfterCheckout:     afterCheckout,
		UndoAuditRecordID: undoAuditRecordID,
		RedoAuditRecordID: redoAuditRecordID,
	}, true
}

func marshalUndoableCheckoutSnapshot(checkout asset.Checkout) (string, error) {
	encoded, err := json.Marshal(undoableCheckoutSnapshot{
		ID:                    checkout.ID.String(),
		TenantID:              checkout.TenantID.String(),
		InventoryID:           checkout.InventoryID.String(),
		AssetID:               checkout.AssetID.String(),
		State:                 checkout.State.String(),
		CheckedOutAt:          checkout.CheckedOutAt.Format(time.RFC3339Nano),
		CheckedOutByPrincipal: checkout.CheckedOutByPrincipal,
		CheckoutDetails:       checkout.CheckoutDetails.String(),
		ReturnedAt:            timeString(checkout.ReturnedAt),
		ReturnedByPrincipal:   checkout.ReturnedByPrincipal,
		ReturnDetails:         checkout.ReturnDetails.String(),
		CreatedAt:             checkout.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:             checkout.UpdatedAt.Format(time.RFC3339Nano),
	})
	return string(encoded), err
}

func unmarshalUndoableCheckoutSnapshot(encoded string) (asset.Checkout, bool) {
	var snapshot undoableCheckoutSnapshot
	if err := json.Unmarshal([]byte(encoded), &snapshot); err != nil {
		return asset.Checkout{}, false
	}
	id, ok := asset.NewCheckoutID(snapshot.ID)
	if !ok {
		return asset.Checkout{}, false
	}
	assetID, ok := asset.NewID(snapshot.AssetID)
	if !ok {
		return asset.Checkout{}, false
	}
	state, ok := asset.NewCheckoutState(snapshot.State)
	if !ok {
		return asset.Checkout{}, false
	}
	checkoutDetails, ok := asset.NewCheckoutDetails(snapshot.CheckoutDetails)
	if !ok {
		return asset.Checkout{}, false
	}
	returnDetails, ok := asset.NewCheckoutDetails(snapshot.ReturnDetails)
	if !ok {
		return asset.Checkout{}, false
	}
	checkedOutAt, ok := parseSnapshotTime(snapshot.CheckedOutAt)
	if !ok {
		return asset.Checkout{}, false
	}
	createdAt, ok := parseSnapshotTime(snapshot.CreatedAt)
	if !ok {
		return asset.Checkout{}, false
	}
	updatedAt, ok := parseSnapshotTime(snapshot.UpdatedAt)
	if !ok {
		return asset.Checkout{}, false
	}
	returnedAt := time.Time{}
	if snapshot.ReturnedAt != "" {
		var parsed bool
		returnedAt, parsed = parseSnapshotTime(snapshot.ReturnedAt)
		if !parsed {
			return asset.Checkout{}, false
		}
	}
	return asset.Checkout{
		ID:                    id,
		TenantID:              asset.TenantID(snapshot.TenantID),
		InventoryID:           asset.InventoryID(snapshot.InventoryID),
		AssetID:               assetID,
		State:                 state,
		CheckedOutAt:          checkedOutAt,
		CheckedOutByPrincipal: snapshot.CheckedOutByPrincipal,
		CheckoutDetails:       checkoutDetails,
		ReturnedAt:            returnedAt,
		ReturnedByPrincipal:   snapshot.ReturnedByPrincipal,
		ReturnDetails:         returnDetails,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
	}, true
}

func timeString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}

func parseSnapshotTime(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return parsed, err == nil
}

func marshalUndoableAssetSnapshot(item asset.Asset, tagIDs []assettag.ID, replacesTags bool) (string, error) {
	encodedTagIDs := make([]string, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		encodedTagIDs = append(encodedTagIDs, tagID.String())
	}
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
		TagIDs:            encodedTagIDs,
		ReplacesTags:      replacesTags,
	})
	return string(encoded), err
}

func unmarshalUndoableAssetSnapshot(encoded string) (asset.Asset, []assettag.ID, bool, bool) {
	var snapshot undoableAssetSnapshot
	if err := json.Unmarshal([]byte(encoded), &snapshot); err != nil {
		return asset.Asset{}, nil, false, false
	}
	id, ok := asset.NewID(snapshot.ID)
	if !ok {
		return asset.Asset{}, nil, false, false
	}
	kind, ok := asset.NewKind(snapshot.Kind)
	if !ok {
		return asset.Asset{}, nil, false, false
	}
	title, ok := asset.NewTitle(snapshot.Title)
	if !ok {
		return asset.Asset{}, nil, false, false
	}
	fields, ok := asset.NewCustomFields(snapshot.CustomFields)
	if !ok {
		return asset.Asset{}, nil, false, false
	}
	tagIDs := make([]assettag.ID, 0, len(snapshot.TagIDs))
	for _, rawTagID := range snapshot.TagIDs {
		tagID, ok := assettag.NewID(rawTagID)
		if !ok {
			return asset.Asset{}, nil, false, false
		}
		tagIDs = append(tagIDs, tagID)
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
	}, tagIDs, snapshot.ReplacesTags, true
}
