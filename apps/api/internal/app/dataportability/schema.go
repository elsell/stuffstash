package dataportability

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s Service) collectSchema(ctx context.Context, in ExportInput, doc *ports.InventoryExportDocument) error {
	doc.Tags = []assettag.Tag{}
	doc.CustomAssetTypes = []customfield.AssetType{}
	doc.CustomFieldDefinitions = []customfield.Definition{}
	doc.Assets = []ports.InventoryExportAsset{}
	err := collectPages(ctx, s.deps.MaxRecords, func(after string, limit int) ([]assettag.Tag, error) {
		return s.deps.Tags.ListAssetTags(ctx, in.TenantID, in.InventoryID, ports.AssetTagPageRequest{IncludeArchived: true, AfterTagID: assettag.ID(after), Limit: limit})
	}, func(t assettag.Tag) string { return t.ID.String() }, func(t assettag.Tag) error {
		if t.TenantID.String() != in.TenantID.String() || t.InventoryID.String() != in.InventoryID.String() {
			return apperrors.ErrUnauthorized
		}
		doc.Tags = append(doc.Tags, t)
		return nil
	})
	if err != nil {
		return err
	}
	err = collectPages(ctx, s.deps.MaxRecords, func(after string, limit int) ([]customfield.AssetType, error) {
		return s.deps.Types.ListInventoryCustomAssetTypes(ctx, in.TenantID, in.InventoryID, ports.CustomAssetTypePageRequest{AfterAssetTypeKey: after, Limit: limit, Lifecycle: ports.CustomizationLifecycleAll})
	}, func(t customfield.AssetType) string { return t.CursorKey() }, func(t customfield.AssetType) error {
		if !validSchemaScope(in, t.TenantID, t.InventoryID, t.Scope) {
			return apperrors.ErrUnauthorized
		}
		doc.CustomAssetTypes = append(doc.CustomAssetTypes, t)
		return nil
	})
	if err != nil {
		return err
	}
	return collectPages(ctx, s.deps.MaxRecords, func(after string, limit int) ([]customfield.Definition, error) {
		return s.deps.Fields.ListInventoryCustomFieldDefinitions(ctx, in.TenantID, in.InventoryID, ports.CustomFieldDefinitionPageRequest{AfterDefinitionKey: after, Limit: limit, Lifecycle: ports.CustomizationLifecycleAll})
	}, func(f customfield.Definition) string { return f.CursorKey() }, func(f customfield.Definition) error {
		if !validSchemaScope(in, f.TenantID, f.InventoryID, f.Scope) {
			return apperrors.ErrUnauthorized
		}
		doc.CustomFieldDefinitions = append(doc.CustomFieldDefinitions, f)
		return nil
	})
}
func validSchemaScope(in ExportInput, t customfield.TenantID, i customfield.InventoryID, scope customfield.Scope) bool {
	return t.String() == in.TenantID.String() && ((scope == customfield.ScopeTenant && i.String() == "") || (scope == customfield.ScopeInventory && i.String() == in.InventoryID.String()))
}
