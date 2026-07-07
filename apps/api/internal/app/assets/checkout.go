package assets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) CheckoutAsset(ctx context.Context, input CheckoutAssetInput) (asset.Checkout, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Checkout{}, err
	}
	if err := s.ensureCheckoutDependencies(); err != nil {
		return asset.Checkout{}, err
	}
	if input.AssetID.String() == "" {
		return asset.Checkout{}, apperrors.ErrInvalidInput
	}
	item, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Checkout{}, err
	}
	if !found {
		return asset.Checkout{}, apperrors.ErrNotFound
	}
	if item.LifecycleState != asset.LifecycleStateActive {
		return asset.Checkout{}, apperrors.ErrInvalidInput
	}
	if _, found, err := s.checkouts.CurrentAssetCheckout(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return asset.Checkout{}, err
	} else if found {
		return asset.Checkout{}, apperrors.ErrInvalidInput
	}
	details, ok := asset.NewCheckoutDetails(input.Details)
	if !ok {
		return asset.Checkout{}, apperrors.ErrInvalidInput
	}
	checkoutID, ok := asset.NewCheckoutID(s.newID())
	if !ok {
		return asset.Checkout{}, apperrors.ErrInvalidInput
	}
	now := s.now().UTC()
	checkout := asset.Checkout{
		ID:                    checkoutID,
		TenantID:              asset.TenantID(input.TenantID.String()),
		InventoryID:           asset.InventoryID(input.InventoryID.String()),
		AssetID:               input.AssetID,
		State:                 asset.CheckoutStateOpen,
		CheckedOutAt:          now,
		CheckedOutByPrincipal: input.Principal.ID.String(),
		CheckoutDetails:       details,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	operation, err := s.newCheckoutUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, audit.ActionAssetCheckedOut, nil, checkout)
	if err != nil {
		return asset.Checkout{}, err
	}
	auditRecord, err := s.newAuditRecord(auditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetCheckedOut,
		TargetType:  audit.TargetAsset,
		TargetID:    input.AssetID.String(),
		Metadata: map[string]string{
			"asset_id":        input.AssetID.String(),
			"checkout_id":     checkout.ID.String(),
			"details_present": boolMetadata(!details.IsEmpty()),
			"operation_id":    operation.ID,
		},
	})
	if err != nil {
		return asset.Checkout{}, err
	}
	if err := s.assetUnitOfWork.CheckOutAsset(ctx, checkout, auditRecord, &operation); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return asset.Checkout{}, apperrors.ErrInvalidInput
		}
		return asset.Checkout{}, err
	}
	s.recordCheckoutEvent(ctx, ports.EventAssetCheckedOut, "asset checked out", checkout, input.Principal.ID.String())
	return checkout, nil
}

func (s Service) ReturnAsset(ctx context.Context, input ReturnAssetInput) (asset.Checkout, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return asset.Checkout{}, err
	}
	if err := s.ensureCheckoutDependencies(); err != nil {
		return asset.Checkout{}, err
	}
	current, found, err := s.checkouts.CurrentAssetCheckout(ctx, input.TenantID, input.InventoryID, input.AssetID)
	if err != nil {
		return asset.Checkout{}, err
	}
	if !found {
		return asset.Checkout{}, apperrors.ErrInvalidInput
	}
	if _, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return asset.Checkout{}, err
	} else if !found {
		return asset.Checkout{}, apperrors.ErrNotFound
	}
	details, ok := asset.NewCheckoutDetails(input.Details)
	if !ok {
		return asset.Checkout{}, apperrors.ErrInvalidInput
	}
	now := s.now().UTC()
	returned := current
	returned.State = asset.CheckoutStateReturned
	returned.ReturnedAt = now
	returned.ReturnedByPrincipal = input.Principal.ID.String()
	returned.ReturnDetails = details
	returned.UpdatedAt = now
	operation, err := s.newCheckoutUndoableOperation(input.Principal.ID, input.Source, input.TenantID, input.InventoryID, audit.ActionAssetReturned, &current, returned)
	if err != nil {
		return asset.Checkout{}, err
	}
	auditRecord, err := s.newAuditRecord(auditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionAssetReturned,
		TargetType:  audit.TargetAsset,
		TargetID:    input.AssetID.String(),
		Metadata: map[string]string{
			"asset_id":                    input.AssetID.String(),
			"checkout_id":                 current.ID.String(),
			"details_present":             boolMetadata(!details.IsEmpty()),
			"checked_out_by_principal_id": current.CheckedOutByPrincipal,
			"operation_id":                operation.ID,
		},
	})
	if err != nil {
		return asset.Checkout{}, err
	}
	if err := s.assetUnitOfWork.ReturnAsset(ctx, current, returned, auditRecord, &operation); err != nil {
		if errors.Is(err, ports.ErrConflict) || errors.Is(err, ports.ErrForbidden) {
			return asset.Checkout{}, apperrors.ErrInvalidInput
		}
		return asset.Checkout{}, err
	}
	s.recordCheckoutEvent(ctx, ports.EventAssetCheckoutReturned, "asset returned", returned, input.Principal.ID.String())
	return returned, nil
}

func (s Service) ListAssetCheckoutHistory(ctx context.Context, input ListAssetCheckoutHistoryInput) (AssetCheckoutHistoryResult, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return AssetCheckoutHistoryResult{}, err
	}
	if s.assets == nil || s.checkouts == nil {
		return AssetCheckoutHistoryResult{}, apperrors.ErrInvalidInput
	}
	if _, found, err := s.assets.AssetByID(ctx, input.TenantID, input.InventoryID, input.AssetID); err != nil {
		return AssetCheckoutHistoryResult{}, err
	} else if !found {
		return AssetCheckoutHistoryResult{}, apperrors.ErrNotFound
	}
	limit := pageLimit(s.defaultPageLimit, s.maxPageLimit, input.Limit)
	cursor, err := decodeCheckoutHistoryCursor(input.Cursor)
	if err != nil {
		return AssetCheckoutHistoryResult{}, err
	}
	items, err := s.checkouts.ListAssetCheckoutHistory(ctx, input.TenantID, input.InventoryID, input.AssetID, ports.AssetCheckoutHistoryPageRequest{
		AfterCheckoutID:   cursor.AfterCheckoutID,
		AfterCheckedOutAt: cursor.AfterCheckedOutAt,
		Limit:             limit + 1,
	})
	if err != nil {
		return AssetCheckoutHistoryResult{}, err
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded := encodeCheckoutHistoryCursor(items[len(items)-1])
		nextCursor = &encoded
	}
	s.observer.Record(ctx, ports.Event{Name: ports.EventAssetCheckoutHistoryListed, Message: "asset checkout history listed", Fields: map[string]string{
		"tenant_id":    input.TenantID.String(),
		"inventory_id": input.InventoryID.String(),
		"asset_id":     input.AssetID.String(),
		"principal_id": input.Principal.ID.String(),
	}})
	return AssetCheckoutHistoryResult{Items: items, Limit: limit, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func (s Service) ListCheckedOutAssets(ctx context.Context, input ListCheckedOutAssetsInput) (CheckedOutAssetsResult, error) {
	if err := s.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return CheckedOutAssetsResult{}, err
	}
	if s.checkouts == nil {
		return CheckedOutAssetsResult{}, apperrors.ErrInvalidInput
	}
	limit := pageLimit(s.defaultPageLimit, s.maxPageLimit, input.Limit)
	cursor, err := decodeCheckedOutAssetsCursor(input.Cursor)
	if err != nil {
		return CheckedOutAssetsResult{}, err
	}
	items, err := s.checkouts.ListCheckedOutAssets(ctx, input.TenantID, input.InventoryID, ports.CheckedOutAssetsPageRequest{
		AfterAssetID:      cursor.AfterAssetID,
		AfterCheckedOutAt: cursor.AfterCheckedOutAt,
		Limit:             limit + 1,
	})
	if err != nil {
		return CheckedOutAssetsResult{}, err
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded := encodeCheckedOutAssetsCursor(items[len(items)-1])
		nextCursor = &encoded
	}
	s.observer.Record(ctx, ports.Event{Name: ports.EventCheckedOutAssetsListed, Message: "checked-out assets listed", Fields: map[string]string{
		"tenant_id":    input.TenantID.String(),
		"inventory_id": input.InventoryID.String(),
		"principal_id": input.Principal.ID.String(),
	}})
	return CheckedOutAssetsResult{Items: items, Limit: limit, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func boolMetadata(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (s Service) recordCheckoutEvent(ctx context.Context, name ports.EventName, message string, checkout asset.Checkout, principalID string) {
	s.observer.Record(ctx, ports.Event{Name: name, Message: message, Fields: map[string]string{
		"tenant_id":    checkout.TenantID.String(),
		"inventory_id": checkout.InventoryID.String(),
		"asset_id":     checkout.AssetID.String(),
		"checkout_id":  checkout.ID.String(),
		"principal_id": principalID,
	}})
}

type checkoutHistoryCursor struct {
	AfterCheckoutID   asset.CheckoutID
	AfterCheckedOutAt time.Time
}

func encodeCheckoutHistoryCursor(checkout asset.Checkout) string {
	encoded, _ := json.Marshal(map[string]string{
		"afterCheckoutId":   checkout.ID.String(),
		"afterCheckedOutAt": checkout.CheckedOutAt.Format(time.RFC3339Nano),
	})
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func decodeCheckoutHistoryCursor(value string) (checkoutHistoryCursor, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return checkoutHistoryCursor{}, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return checkoutHistoryCursor{}, apperrors.ErrInvalidInput
	}
	var payload map[string]string
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return checkoutHistoryCursor{}, apperrors.ErrInvalidInput
	}
	checkoutID, ok := asset.NewCheckoutID(payload["afterCheckoutId"])
	if !ok {
		return checkoutHistoryCursor{}, apperrors.ErrInvalidInput
	}
	checkedOutAt, err := time.Parse(time.RFC3339Nano, payload["afterCheckedOutAt"])
	if err != nil {
		return checkoutHistoryCursor{}, apperrors.ErrInvalidInput
	}
	return checkoutHistoryCursor{AfterCheckoutID: checkoutID, AfterCheckedOutAt: checkedOutAt}, nil
}

type checkedOutAssetsCursor struct {
	AfterAssetID      asset.ID
	AfterCheckedOutAt time.Time
}

func encodeCheckedOutAssetsCursor(item ports.CheckedOutAsset) string {
	encoded, _ := json.Marshal(map[string]string{
		"afterAssetId":      item.Asset.ID.String(),
		"afterCheckedOutAt": item.Checkout.CheckedOutAt.Format(time.RFC3339Nano),
	})
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func decodeCheckedOutAssetsCursor(value string) (checkedOutAssetsCursor, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return checkedOutAssetsCursor{}, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return checkedOutAssetsCursor{}, apperrors.ErrInvalidInput
	}
	var payload map[string]string
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return checkedOutAssetsCursor{}, apperrors.ErrInvalidInput
	}
	assetID, ok := asset.NewID(payload["afterAssetId"])
	if !ok {
		return checkedOutAssetsCursor{}, apperrors.ErrInvalidInput
	}
	checkedOutAt, err := time.Parse(time.RFC3339Nano, payload["afterCheckedOutAt"])
	if err != nil {
		return checkedOutAssetsCursor{}, apperrors.ErrInvalidInput
	}
	return checkedOutAssetsCursor{AfterAssetID: assetID, AfterCheckedOutAt: checkedOutAt}, nil
}
