package dataportability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func BuildArchiveRestorePlan(ctx context.Context, source ports.InventoryExportDocument, target ports.ArchiveRestoreDestination, reserved ports.ArchiveKeyReservations, selected ports.ArchiveMediaSelection, ids ports.IDGenerator, maxRecords int) (ports.ArchiveRestorePlan, error) {
	empty := ports.ArchiveRestorePlan{}
	if err := ValidateArchiveDocument(ctx, source, maxRecords); err != nil {
		return empty, err
	}
	if ids == nil || !archiveID(target.TenantID) || !archiveID(target.InventoryID) || !archiveID(target.PrincipalID) || target.InventoryID == source.InventoryID {
		return empty, ErrArchiveMetadata
	}
	if _, ok := inventory.NewName(target.Name); !ok {
		return empty, ErrArchiveMetadata
	}
	remap, err := allocateArchiveIDs(ctx, source, target, ids)
	if err != nil {
		return empty, err
	}
	result := ports.ArchiveRestorePlan{Document: ports.InventoryExportDocument{SchemaVersion: 2, ExportedAt: source.ExportedAt, TenantID: target.TenantID, InventoryID: target.InventoryID, InventoryName: target.Name}, CheckoutSourcePrincipals: map[string]string{}}
	typeKeys := copyReservations(reserved.Types)
	fieldKeys := copyReservations(reserved.Fields)
	// Reserve every non-conflicting source key before generating alternatives.
	for _, t := range source.CustomAssetTypes {
		typeKeys[t.Key.String()] = true
	}
	for _, f := range source.CustomFieldDefinitions {
		fieldKeys[f.Key.String()] = true
	}
	for _, t := range source.CustomAssetTypes {
		key := restoreKey("type", t.ID.String(), t.Key.String(), reserved.Types, typeKeys)
		if key != t.Key.String() {
			result.KeyRemappings = append(result.KeyRemappings, ports.ArchiveKeyRemapping{Family: "type", SourceKey: t.Key.String(), DestinationKey: key})
		}
		t.ID = customfield.AssetTypeID(remap.get("type", t.ID.String()))
		t.Key = customfield.Key(key)
		t.Scope = customfield.ScopeInventory
		t.TenantID = customfield.TenantID(target.TenantID)
		t.InventoryID = customfield.InventoryID(target.InventoryID)
		result.Document.CustomAssetTypes = append(result.Document.CustomAssetTypes, t)
	}
	fieldNames := map[string]string{}
	for _, f := range source.CustomFieldDefinitions {
		key := restoreKey("field", f.ID.String(), f.Key.String(), reserved.Fields, fieldKeys)
		if key != f.Key.String() {
			result.KeyRemappings = append(result.KeyRemappings, ports.ArchiveKeyRemapping{Family: "field", SourceKey: f.Key.String(), DestinationKey: key})
		}
		fieldNames[f.Key.String()] = key
		f.ID = customfield.ID(remap.get("field", f.ID.String()))
		f.Key = customfield.Key(key)
		f.Scope = customfield.ScopeInventory
		f.TenantID = customfield.TenantID(target.TenantID)
		f.InventoryID = customfield.InventoryID(target.InventoryID)
		f.EnumOptions = append([]customfield.Key(nil), f.EnumOptions...)
		targets := make([]customfield.AssetTypeID, len(f.CustomAssetTypeIDs))
		for i, id := range f.CustomAssetTypeIDs {
			targets[i] = customfield.AssetTypeID(remap.get("type", id.String()))
		}
		f.CustomAssetTypeIDs = targets
		result.Document.CustomFieldDefinitions = append(result.Document.CustomFieldDefinitions, f)
	}
	for _, t := range source.Tags {
		t.ID = assettag.ID(remap.get("tag", t.ID.String()))
		t.TenantID = assettag.TenantID(target.TenantID)
		t.InventoryID = assettag.InventoryID(target.InventoryID)
		result.Document.Tags = append(result.Document.Tags, t)
	}
	for _, a := range source.Assets {
		if err := ctx.Err(); err != nil {
			return empty, err
		}
		a.ID = remap.get("asset", a.ID)
		a.ParentAssetID = remap.get("asset", a.ParentAssetID)
		a.CustomAssetTypeID = remap.get("type", a.CustomAssetTypeID)
		fields := map[string]any{}
		for key, value := range a.CustomFields {
			fields[fieldNames[key]] = value
		}
		a.CustomFields = fields
		tags := make([]string, len(a.TagIDs))
		for i, id := range a.TagIDs {
			tags[i] = remap.get("tag", id)
		}
		a.TagIDs = tags
		attachments := []media.Attachment{}
		for _, m := range a.Attachments {
			include := selected.OtherFiles
			if m.ContentType.IsImage() {
				include = selected.Photos
			}
			if !include {
				result.OmittedAttachments++
				continue
			}
			m.ID = media.ID(remap.get("attachment", m.ID.String()))
			m.TenantID = media.TenantID(target.TenantID)
			m.InventoryID = media.InventoryID(target.InventoryID)
			m.AssetID = media.AssetID(a.ID)
			key, ok := media.NewStorageKey(target.TenantID + "/" + target.InventoryID + "/" + a.ID + "/" + m.ID.String())
			if !ok {
				return empty, ErrArchiveMetadata
			}
			m.StorageKey = key
			attachments = append(attachments, m)
		}
		a.Attachments = attachments
		if a.CurrentCheckout != nil {
			c := *a.CurrentCheckout
			c.ID = asset.CheckoutID(remap.get("checkout", c.ID.String()))
			c.TenantID = asset.TenantID(target.TenantID)
			c.InventoryID = asset.InventoryID(target.InventoryID)
			c.AssetID = asset.ID(a.ID)
			result.CheckoutSourcePrincipals[c.ID.String()] = c.CheckedOutByPrincipal
			c.CheckedOutByPrincipal = target.PrincipalID
			a.CurrentCheckout = &c
		}
		result.Document.Assets = append(result.Document.Assets, a)
	}
	if err := ValidateArchiveDocument(ctx, result.Document, maxRecords); err != nil {
		return empty, err
	}
	return result, nil
}

func copyReservations(source map[string]bool) map[string]bool {
	result := map[string]bool{}
	for key, reserved := range source {
		if reserved {
			result[key] = true
		}
	}
	return result
}
func restoreKey(family, id, key string, reservations, used map[string]bool) string {
	if !reservations[key] {
		return key
	}
	digest := sha256.Sum256([]byte(family + ":" + id))
	base := "restored-" + hex.EncodeToString(digest[:10])
	candidate := base
	for suffix := 1; used[candidate]; suffix++ {
		candidate = base + "-" + strconv.Itoa(suffix)
	}
	used[candidate] = true
	return candidate
}

type archiveIDMapping map[string]string

func (m archiveIDMapping) get(family, id string) string {
	if id == "" {
		return ""
	}
	return m[family+":"+id]
}
func allocateArchiveIDs(ctx context.Context, d ports.InventoryExportDocument, target ports.ArchiveRestoreDestination, ids ports.IDGenerator) (archiveIDMapping, error) {
	type reference struct{ family, id string }
	sources := []reference{}
	for _, t := range d.Tags {
		sources = append(sources, reference{"tag", t.ID.String()})
	}
	for _, t := range d.CustomAssetTypes {
		sources = append(sources, reference{"type", t.ID.String()})
	}
	for _, f := range d.CustomFieldDefinitions {
		sources = append(sources, reference{"field", f.ID.String()})
	}
	for _, a := range d.Assets {
		sources = append(sources, reference{"asset", a.ID})
		for _, m := range a.Attachments {
			sources = append(sources, reference{"attachment", m.ID.String()})
		}
		if a.CurrentCheckout != nil {
			sources = append(sources, reference{"checkout", a.CurrentCheckout.ID.String()})
		}
	}
	used := map[string]bool{d.InventoryID: true, d.TenantID: true}
	used[target.InventoryID] = true
	used[target.TenantID] = true
	for _, ref := range sources {
		used[ref.id] = true
	}
	result := archiveIDMapping{}
	for _, ref := range sources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		next := ids.NewID()
		// A broken generator must fail rather than alias two restored resources.
		if !archiveID(next) || used[next] {
			return nil, ErrArchiveMetadata
		}
		used[next] = true
		result[ref.family+":"+ref.id] = next
	}
	return result, nil
}
