package assets

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) newAssetUndoableOperation(principalID identity.PrincipalID, source audit.Source, tenantID tenant.ID, inventoryID inventory.InventoryID, originalAction audit.Action, before *asset.Asset, after asset.Asset) (ports.UndoableOperation, error) {
	if s.undoables == nil {
		return ports.UndoableOperation{}, apperrors.ErrInvalidInput
	}
	id := s.newID()
	if strings.TrimSpace(id) == "" {
		return ports.UndoableOperation{}, apperrors.ErrInvalidInput
	}
	var beforeCopy *asset.Asset
	if before != nil {
		copied := *before
		beforeCopy = &copied
	}
	return ports.UndoableOperation{
		ID:             id,
		TenantID:       tenantID,
		InventoryID:    inventoryID,
		PrincipalID:    principalID,
		Source:         source,
		TargetType:     audit.TargetAsset,
		TargetID:       after.ID.String(),
		OriginalAction: originalAction,
		Status:         ports.UndoableOperationAvailable,
		CreatedAt:      s.now().UTC(),
		BeforeAsset:    beforeCopy,
		AfterAsset:     after,
	}, nil
}

func (s Service) newCheckoutUndoableOperation(principalID identity.PrincipalID, source audit.Source, tenantID tenant.ID, inventoryID inventory.InventoryID, originalAction audit.Action, before *asset.Checkout, after asset.Checkout) (ports.UndoableOperation, error) {
	if s.undoables == nil {
		return ports.UndoableOperation{}, apperrors.ErrInvalidInput
	}
	id := s.newID()
	if strings.TrimSpace(id) == "" {
		return ports.UndoableOperation{}, apperrors.ErrInvalidInput
	}
	var beforeCopy *asset.Checkout
	if before != nil {
		copied := *before
		beforeCopy = &copied
	}
	afterCopy := after
	return ports.UndoableOperation{
		ID:             id,
		TenantID:       tenantID,
		InventoryID:    inventoryID,
		PrincipalID:    principalID,
		Source:         source,
		TargetType:     audit.TargetAsset,
		TargetID:       after.AssetID.String(),
		OriginalAction: originalAction,
		Status:         ports.UndoableOperationAvailable,
		CreatedAt:      s.now().UTC(),
		BeforeCheckout: beforeCopy,
		AfterCheckout:  &afterCopy,
	}, nil
}
