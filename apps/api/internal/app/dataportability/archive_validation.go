package dataportability

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var ErrArchiveMetadata = errors.New("archive inventory metadata is invalid")

// ValidateArchiveDocument validates source meaning before preview or restore.
// Source identifiers remain references, never destination authority.
func ValidateArchiveDocument(ctx context.Context, d ports.InventoryExportDocument, maxRecords int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if maxRecords <= 0 {
		return ports.ErrInventoryExportLimit
	}
	if d.SchemaVersion != 2 || !archiveID(d.TenantID) || !archiveID(d.InventoryID) || d.ExportedAt.IsZero() {
		return ErrArchiveMetadata
	}
	if _, ok := inventory.NewName(d.InventoryName); !ok {
		return ErrArchiveMetadata
	}
	count := len(d.Assets) + len(d.Tags) + len(d.CustomAssetTypes) + len(d.CustomFieldDefinitions)
	if count > maxRecords {
		return ports.ErrInventoryExportLimit
	}
	if err := validateArchiveSchema(d); err != nil {
		return err
	}
	tags := map[string]bool{}
	tagKeys := map[string]bool{}
	for _, t := range d.Tags {
		if !archiveID(t.ID.String()) || tags[t.ID.String()] || tagKeys[t.Key.String()] || !archiveLifecycle(t.LifecycleState.String()) || t.CreatedAt.IsZero() || t.UpdatedAt.Before(t.CreatedAt) {
			return ErrArchiveMetadata
		}
		if key, ok := assettag.NewKey(t.Key.String()); !ok || key != t.Key {
			return ErrArchiveMetadata
		}
		if _, ok := assettag.NewDisplayName(t.DisplayName.String()); !ok {
			return ErrArchiveMetadata
		}
		if _, ok := assettag.NewColor(t.Color.String()); !ok {
			return ErrArchiveMetadata
		}
		tags[t.ID.String()] = true
		tagKeys[t.Key.String()] = true
	}
	types := map[string]bool{}
	for _, t := range d.CustomAssetTypes {
		types[t.ID.String()] = true
	}
	assets := make(map[string]ports.InventoryExportAsset, len(d.Assets))
	attachments := map[string]bool{}
	checkouts := map[string]bool{}
	for _, a := range d.Assets {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, exists := assets[a.ID]; exists || !archiveID(a.ID) || a.CreatedAt.IsZero() || a.UpdatedAt.Before(a.CreatedAt) || !archiveLifecycle(a.LifecycleState) {
			return ErrArchiveMetadata
		}
		kind, ok := asset.NewKind(a.Kind)
		if !ok || kind.String() != a.Kind {
			return ErrArchiveMetadata
		}
		if _, ok := asset.NewTitle(a.Title); !ok {
			return ErrArchiveMetadata
		}
		if a.CustomAssetTypeID != "" && !types[a.CustomAssetTypeID] {
			return ErrArchiveMetadata
		}
		for rawKey := range a.CustomFields {
			key, ok := customfield.NewKey(rawKey)
			if !ok || key.String() != rawKey {
				return ErrArchiveMetadata
			}
		}
		if !customfield.DefinitionSet(d.CustomFieldDefinitions).ValidateValuesForAssetType(a.CustomFields, customfield.AssetTypeID(a.CustomAssetTypeID)) {
			return ErrArchiveMetadata
		}
		assigned := map[string]bool{}
		for _, id := range a.TagIDs {
			if !tags[id] || assigned[id] {
				return ErrArchiveMetadata
			}
			assigned[id] = true
		}
		if a.ExpirationDate != "" || a.ExpirationPrecision != "" {
			if _, err := expirationdate.ParseDate(a.ExpirationDate, expirationdate.Precision(a.ExpirationPrecision)); err != nil {
				return ErrArchiveMetadata
			}
		}
		count += len(a.Attachments)
		if a.CurrentCheckout != nil {
			count++
		}
		if count > maxRecords {
			return ports.ErrInventoryExportLimit
		}
		if err := validateArchiveMedia(a, attachments); err != nil {
			return err
		}
		if c := a.CurrentCheckout; c != nil {
			if !kind.IsPortable() || !archiveID(c.ID.String()) || checkouts[c.ID.String()] || c.State != asset.CheckoutStateOpen || c.CheckedOutAt.IsZero() || c.CreatedAt.IsZero() || c.UpdatedAt.Before(c.CreatedAt) || strings.TrimSpace(c.CheckedOutByPrincipal) == "" {
				return ErrArchiveMetadata
			}
			if _, ok := asset.NewCheckoutDetails(c.CheckoutDetails.String()); !ok {
				return ErrArchiveMetadata
			}
			checkouts[c.ID.String()] = true
		}
		assets[a.ID] = a
	}
	return validateArchiveParents(ctx, assets)
}
func archiveID(id string) bool           { return id != "" && strings.TrimSpace(id) == id }
func archiveLifecycle(state string) bool { return state == "active" || state == "archived" }

func validateArchiveParents(ctx context.Context, assets map[string]ports.InventoryExportAsset) error {
	// Iterative color walk handles deeply nested inventories without recursion and
	// visits each edge at most twice.
	colors := map[string]uint8{}
	for id := range assets {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := []string{}
		for id != "" && colors[id] != 2 {
			if colors[id] == 1 {
				return ErrArchiveMetadata
			}
			a, exists := assets[id]
			if !exists {
				return ErrArchiveMetadata
			}
			colors[id] = 1
			path = append(path, id)
			if a.ParentAssetID != "" {
				parent, exists := assets[a.ParentAssetID]
				if !exists || !asset.Kind(parent.Kind).CanContainChildren() {
					return ErrArchiveMetadata
				}
			}
			id = a.ParentAssetID
		}
		for _, id := range path {
			colors[id] = 2
		}
	}
	return nil
}
