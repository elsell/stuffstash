// Package dataportability orchestrates authorized inventory transfers through ports.
package dataportability

import (
	"context"
	"errors"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Dependencies struct {
	Authorizer           ports.Authorizer
	Inventories          ports.InventoryRepository
	Assets               ports.AssetRepository
	Tags                 ports.AssetTagRepository
	Checkouts            ports.AssetCheckoutRepository
	Attachments          ports.AttachmentRepository
	Types                ports.CustomAssetTypeRepository
	Fields               ports.CustomFieldDefinitionRepository
	Audit                ports.AuditRepository
	IDs                  ports.IDGenerator
	Clock                ports.Clock
	Observer             ports.Observer
	Encoder              ports.InventoryExportEncoder
	MaxRecords, MaxBytes int
}
type Service struct{ deps Dependencies }

func New(deps Dependencies) Service { return Service{deps: deps} }

type ExportInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	RequestID   string
	Source      audit.Source
	Format      ports.InventoryExportFormat
}
type ExportResult struct {
	Content []byte
	Format  ports.InventoryExportFormat
}

func (s Service) Export(ctx context.Context, in ExportInput) (ExportResult, error) {
	empty := ExportResult{}
	d := s.deps
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	if in.Format == "" {
		in.Format = ports.InventoryExportJSON
	}
	if in.Format != ports.InventoryExportJSON && in.Format != ports.InventoryExportCSV {
		return empty, ports.ErrInventoryExportFormat
	}
	if d.Authorizer == nil || d.Inventories == nil || d.Assets == nil || d.Tags == nil || d.Checkouts == nil || d.Attachments == nil || d.Types == nil || d.Fields == nil || d.Audit == nil || d.IDs == nil || d.Clock == nil || d.Encoder == nil || d.MaxRecords <= 0 || d.MaxBytes <= 0 {
		return empty, errors.New("inventory export is not configured")
	}
	item, found, err := d.Inventories.InventoryByID(ctx, in.TenantID, in.InventoryID)
	if err != nil {
		return empty, err
	}
	if !found || item.TenantID.String() != in.TenantID.String() || item.ID != in.InventoryID {
		return empty, apperrors.ErrNotFound
	}
	if err = d.Authorizer.CheckInventory(ctx, in.Principal, ports.InventoryPermissionView, in.InventoryID); err != nil {
		return empty, err
	}
	document := ports.InventoryExportDocument{SchemaVersion: 1, ExportedAt: d.Clock.Now(), TenantID: in.TenantID.String(), InventoryID: in.InventoryID.String(), InventoryName: item.Name.String()}
	if err = s.collectSchema(ctx, in, &document); err != nil {
		return empty, err
	}
	if err = s.collectAssets(ctx, in, &document); err != nil {
		return empty, err
	}
	body, err := d.Encoder.Encode(ctx, document, in.Format, d.MaxBytes)
	if err != nil {
		return empty, err
	}
	if err = d.Authorizer.CheckInventory(ctx, in.Principal, ports.InventoryPermissionView, in.InventoryID); err != nil {
		return empty, err
	}
	metadata := map[string]string{"format": string(in.Format), "asset_count": strconv.Itoa(len(document.Assets))}
	if err = appsupport.SaveReadAuditRecord(ctx, d.Audit, d.IDs, d.Clock, appsupport.AuditRecordInput{Principal: in.Principal, TenantID: in.TenantID, InventoryID: in.InventoryID, RequestID: in.RequestID, Source: in.Source, Action: audit.ActionInventoryExported, TargetType: audit.TargetInventory, TargetID: in.InventoryID.String(), Metadata: metadata}); err != nil {
		return empty, err
	}
	if err = ctx.Err(); err != nil {
		return empty, err
	}
	if d.Observer != nil {
		d.Observer.Record(ctx, ports.Event{Name: ports.EventInventoryExported, Message: "inventory exported", Fields: metadata})
	}
	return ExportResult{Content: body, Format: in.Format}, nil
}

// collectPages detects repeated/out-of-order data instead of silently publishing a
// truncated file. Read one extra row at the bound to distinguish full from partial.
func collectPages[T any](ctx context.Context, limit int, fetch func(string, int) ([]T, error), key func(T) string, accept func(T) error) error {
	cursor := ""
	count := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		size := min(200, limit-count+1)
		page, err := fetch(cursor, size)
		if err != nil {
			return err
		}
		if len(page) == 0 {
			return nil
		}
		for _, value := range page {
			if err := ctx.Err(); err != nil {
				return err
			}
			next := key(value)
			if next <= cursor {
				return errors.New("inventory export repository returned a nonadvancing page")
			}
			count++
			if count > limit {
				return ports.ErrInventoryExportLimit
			}
			if err := accept(value); err != nil {
				return err
			}
			cursor = next
		}
	}
}
