package gormstore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveAuditRecord(ctx context.Context, record audit.Record) error {
	return createAuditRecord(s.db.WithContext(ctx), record)
}

func createAuditRecord(tx *gorm.DB, record audit.Record) error {
	metadata, err := json.Marshal(record.MetadataValues())
	if err != nil {
		return err
	}
	model := auditRecordModel{
		ID:          record.ID.String(),
		TenantID:    record.TenantID.String(),
		PrincipalID: record.PrincipalID.String(),
		Action:      record.Action.String(),
		Source:      record.Source.String(),
		TargetType:  record.TargetType.String(),
		TargetID:    record.TargetID,
		OccurredAt:  record.OccurredAt,
		RequestID:   record.RequestID,
		Metadata:    string(metadata),
	}
	if record.InventoryID.String() != "" {
		inventoryID := record.InventoryID.String()
		model.InventoryID = &inventoryID
	}
	return tx.Create(&model).Error
}

func (s Store) ListTenantAuditRecords(ctx context.Context, tenantID tenant.ID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	query := s.db.WithContext(ctx).Where(&auditRecordModel{TenantID: tenantID.String()})
	query = query.Where(clause.Eq{Column: clause.Column{Name: "inventory_id"}, Value: nil})
	return s.listAuditRecords(query, page)
}

func (s Store) ListInventoryAuditRecords(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	query := s.db.WithContext(ctx).Where(&auditRecordModel{
		TenantID: tenantID.String(),
	})
	query = query.Where(&auditRecordModel{InventoryID: stringPtrFromInventoryID(inventoryID)})
	return s.listAuditRecords(query, page)
}

func (s Store) ListAssetAuditRecords(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, targetID string, request ports.AssetAuditRecordListRequest) ([]audit.Record, error) {
	query := s.db.WithContext(ctx).Where(&auditRecordModel{
		TenantID:   tenantID.String(),
		TargetType: audit.TargetAsset.String(),
		TargetID:   targetID,
	})
	query = query.Where(&auditRecordModel{InventoryID: stringPtrFromInventoryID(inventoryID)})
	if len(request.Actions) > 0 {
		actions := make([]string, 0, len(request.Actions))
		for _, action := range request.Actions {
			actions = append(actions, action.String())
		}
		query = query.Where(clause.IN{Column: clause.Column{Name: "action"}, Values: stringValues(actions)})
	}
	if !request.BeforeOccurredAt.IsZero() && request.BeforeRecordID.String() != "" {
		query = query.Where(clause.Or(
			clause.Lt{Column: clause.Column{Name: "occurred_at"}, Value: request.BeforeOccurredAt},
			clause.And(
				clause.Eq{Column: clause.Column{Name: "occurred_at"}, Value: request.BeforeOccurredAt},
				clause.Lt{Column: clause.Column{Name: "id"}, Value: request.BeforeRecordID.String()},
			),
		))
	}
	var models []auditRecordModel
	if request.Limit > 0 {
		query = query.Limit(request.Limit)
	}
	if err := query.Order(clause.OrderBy{
		Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "occurred_at"}, Desc: true},
			{Column: clause.Column{Name: "id"}, Desc: true},
		},
	}).Find(&models).Error; err != nil {
		return nil, err
	}
	return auditRecordModelsToDomain(models)
}

func (s Store) listAuditRecords(query *gorm.DB, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	var models []auditRecordModel
	if !page.AfterOccurredAt.IsZero() && page.AfterRecordID.String() != "" {
		query = query.Where(clause.Or(
			clause.Gt{Column: clause.Column{Name: "occurred_at"}, Value: page.AfterOccurredAt},
			clause.And(
				clause.Eq{Column: clause.Column{Name: "occurred_at"}, Value: page.AfterOccurredAt},
				clause.Gt{Column: clause.Column{Name: "id"}, Value: page.AfterRecordID.String()},
			),
		))
	}
	if page.Limit > 0 {
		query = query.Limit(page.Limit)
	}
	if err := query.Order(clause.OrderBy{
		Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "occurred_at"}},
			{Column: clause.Column{Name: "id"}},
		},
	}).Find(&models).Error; err != nil {
		return nil, err
	}

	return auditRecordModelsToDomain(models)
}

func auditRecordModelsToDomain(models []auditRecordModel) ([]audit.Record, error) {
	items := make([]audit.Record, 0, len(models))
	for _, model := range models {
		item, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid audit record row %q", model.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func stringPtrFromInventoryID(id inventory.InventoryID) *string {
	if id.String() == "" {
		return nil
	}
	value := id.String()
	return &value
}
