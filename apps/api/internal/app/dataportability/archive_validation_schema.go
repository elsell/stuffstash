package dataportability

import (
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func validateArchiveSchema(d ports.InventoryExportDocument) error {
	types := map[string]bool{}
	typeKeys := map[string]bool{}
	for _, t := range d.CustomAssetTypes {
		if !archiveID(t.ID.String()) || types[t.ID.String()] || typeKeys[t.Key.String()] || !archiveLifecycle(t.LifecycleState.String()) || !archiveScope(t.Scope) {
			return ErrArchiveMetadata
		}
		if key, ok := customfield.NewKey(t.Key.String()); !ok || key != t.Key {
			return ErrArchiveMetadata
		}
		if _, ok := customfield.NewDisplayName(t.DisplayName.String()); !ok {
			return ErrArchiveMetadata
		}
		if _, ok := customfield.NewDescription(t.Description.String()); !ok {
			return ErrArchiveMetadata
		}
		types[t.ID.String()] = true
		typeKeys[t.Key.String()] = true
	}
	fields := map[string]bool{}
	fieldKeys := map[string]bool{}
	for _, f := range d.CustomFieldDefinitions {
		if !archiveID(f.ID.String()) || fields[f.ID.String()] || fieldKeys[f.Key.String()] || !archiveLifecycle(f.LifecycleState.String()) || !archiveScope(f.Scope) {
			return ErrArchiveMetadata
		}
		if key, ok := customfield.NewKey(f.Key.String()); !ok || key != f.Key {
			return ErrArchiveMetadata
		}
		if _, ok := customfield.NewDisplayName(f.DisplayName.String()); !ok {
			return ErrArchiveMetadata
		}
		if _, ok := customfield.NewFieldType(f.Type.String()); !ok {
			return ErrArchiveMetadata
		}
		for _, option := range f.EnumOptions {
			if key, ok := customfield.NewKey(option.String()); !ok || key != option {
				return ErrArchiveMetadata
			}
		}
		for _, id := range f.CustomAssetTypeIDs {
			if !types[id.String()] {
				return ErrArchiveMetadata
			}
		}
		// Validate the definition structure at local scope, which is its restore scope.
		if _, ok := customfield.NewDefinitionWithLifecycle(f.ID, customfield.TenantID(d.TenantID), customfield.InventoryID(d.InventoryID), customfield.ScopeInventory, f.Key, f.DisplayName, f.Type, f.EnumOptions, f.Applicability, f.CustomAssetTypeIDs, f.LifecycleState); !ok {
			return ErrArchiveMetadata
		}
		fields[f.ID.String()] = true
		fieldKeys[f.Key.String()] = true
	}
	return nil
}
func archiveScope(scope customfield.Scope) bool {
	return scope == customfield.ScopeTenant || scope == customfield.ScopeInventory
}
func validateArchiveMedia(a ports.InventoryExportAsset, seen map[string]bool) error {
	for _, m := range a.Attachments {
		if !archiveID(m.ID.String()) || seen[m.ID.String()] || m.SizeBytes <= 0 || m.CreatedAt.IsZero() || !archiveLifecycle(m.LifecycleState.String()) {
			return ErrArchiveMetadata
		}
		if _, ok := media.NewFileName(m.FileName.String()); !ok {
			return ErrArchiveMetadata
		}
		if c, ok := media.NewContentType(m.ContentType.String()); !ok || c != m.ContentType {
			return ErrArchiveMetadata
		}
		if h, ok := media.NewSHA256(m.SHA256.String()); !ok || h != m.SHA256 {
			return ErrArchiveMetadata
		}
		seen[m.ID.String()] = true
	}
	return nil
}
