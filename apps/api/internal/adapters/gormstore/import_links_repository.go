package gormstore

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) ImportSourceLinkByKey(ctx context.Context, key ports.ImportSourceLinkKey) (ports.ImportSourceLink, bool, error) {
	if err := validateImportSourceLinkKey(key); err != nil {
		return ports.ImportSourceLink{}, false, err
	}
	var model importSourceLinkModel
	err := s.db.WithContext(ctx).
		Where(&importSourceLinkModel{
			TenantID:          key.TenantID.String(),
			InventoryID:       key.InventoryID.String(),
			SourceType:        string(key.SourceType),
			SourceInstanceKey: strings.TrimSpace(key.SourceInstanceKey),
			SourceEntityType:  string(key.SourceEntityType),
			SourceEntityID:    strings.TrimSpace(key.SourceEntityID),
		}).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ImportSourceLink{}, false, nil
	}
	if err != nil {
		return ports.ImportSourceLink{}, false, err
	}
	return importSourceLinkFromModel(model), true, nil
}

func (s Store) SaveImportSourceLink(ctx context.Context, link ports.ImportSourceLink) error {
	if err := validateImportSourceLink(link); err != nil {
		return err
	}
	model := importSourceLinkModelFromRecord(link)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return importLinkPersistenceError(err)
	}
	return nil
}

func (s Store) SaveImportJobResource(ctx context.Context, record ports.ImportJobResource) error {
	if err := validateImportJobResource(record); err != nil {
		return err
	}
	model := importJobResourceModelFromRecord(record)
	if err := s.db.WithContext(ctx).Create(&model).Error; err != nil {
		return importLinkPersistenceError(err)
	}
	return nil
}

func (s Store) ListImportJobResources(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, page ports.ImportJobResourcePageRequest) ([]ports.ImportJobResource, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" {
		return nil, ports.ErrInvalidProviderInput
	}
	limit := page.Limit
	if limit <= 0 {
		limit = 50
	}
	var models []importJobResourceModel
	if err := s.db.WithContext(ctx).
		Where(&importJobResourceModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), JobID: jobID.String()}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "resource_type"}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "resource_id"}}).
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}
	return importJobResourceRecordsFromModels(models), nil
}

func (s Store) ListAllImportJobResources(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) ([]ports.ImportJobResource, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" {
		return nil, ports.ErrInvalidProviderInput
	}
	var models []importJobResourceModel
	if err := s.db.WithContext(ctx).
		Where(&importJobResourceModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), JobID: jobID.String()}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "resource_type"}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "resource_id"}}).
		Find(&models).Error; err != nil {
		return nil, err
	}
	return importJobResourceRecordsFromModels(models), nil
}

func importJobResourceRecordsFromModels(models []importJobResourceModel) []ports.ImportJobResource {
	records := make([]ports.ImportJobResource, 0, len(models))
	for _, model := range models {
		records = append(records, importJobResourceFromModel(model))
	}
	return records
}

func (s Store) DeleteImportSourceLinksForJob(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (int, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" {
		return 0, ports.ErrInvalidProviderInput
	}
	result := s.db.WithContext(ctx).
		Where(&importSourceLinkModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), JobID: jobID.String()}).
		Delete(&importSourceLinkModel{})
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

func importSourceLinkModelFromRecord(link ports.ImportSourceLink) importSourceLinkModel {
	return importSourceLinkModel{
		TenantID:          link.Key.TenantID.String(),
		InventoryID:       link.Key.InventoryID.String(),
		JobID:             link.JobID.String(),
		SourceType:        string(link.Key.SourceType),
		SourceInstanceKey: strings.TrimSpace(link.Key.SourceInstanceKey),
		SourceEntityType:  string(link.Key.SourceEntityType),
		SourceEntityID:    strings.TrimSpace(link.Key.SourceEntityID),
		ResourceType:      string(link.ResourceType),
		ResourceID:        strings.TrimSpace(link.ResourceID),
		CreatedAt:         link.CreatedAt,
	}
}

func importSourceLinkFromModel(model importSourceLinkModel) ports.ImportSourceLink {
	return ports.ImportSourceLink{
		Key: ports.ImportSourceLinkKey{
			TenantID:          tenant.ID(model.TenantID),
			InventoryID:       inventory.InventoryID(model.InventoryID),
			SourceType:        importplan.SourceType(model.SourceType),
			SourceInstanceKey: model.SourceInstanceKey,
			SourceEntityType:  ports.ImportSourceEntityType(model.SourceEntityType),
			SourceEntityID:    model.SourceEntityID,
		},
		ResourceType: ports.ImportResourceType(model.ResourceType),
		ResourceID:   model.ResourceID,
		JobID:        importjob.ID(model.JobID),
		CreatedAt:    model.CreatedAt,
	}
}

func importJobResourceModelFromRecord(record ports.ImportJobResource) importJobResourceModel {
	return importJobResourceModel{
		TenantID:          record.TenantID.String(),
		InventoryID:       record.InventoryID.String(),
		JobID:             record.JobID.String(),
		ResourceType:      string(record.ResourceType),
		ResourceID:        strings.TrimSpace(record.ResourceID),
		ResourceOwnerID:   strings.TrimSpace(record.ResourceOwnerID),
		SourceType:        string(record.SourceType),
		SourceInstanceKey: strings.TrimSpace(record.SourceInstanceKey),
		SourceEntityType:  string(record.SourceEntityType),
		SourceEntityID:    strings.TrimSpace(record.SourceEntityID),
		CreatedAt:         record.CreatedAt,
	}
}

func importJobResourceFromModel(model importJobResourceModel) ports.ImportJobResource {
	return ports.ImportJobResource{
		TenantID:          tenant.ID(model.TenantID),
		InventoryID:       inventory.InventoryID(model.InventoryID),
		JobID:             importjob.ID(model.JobID),
		ResourceType:      ports.ImportResourceType(model.ResourceType),
		ResourceID:        model.ResourceID,
		ResourceOwnerID:   model.ResourceOwnerID,
		SourceType:        importplan.SourceType(model.SourceType),
		SourceInstanceKey: model.SourceInstanceKey,
		SourceEntityType:  ports.ImportSourceEntityType(model.SourceEntityType),
		SourceEntityID:    model.SourceEntityID,
		CreatedAt:         model.CreatedAt,
	}
}

func validateImportSourceLink(link ports.ImportSourceLink) error {
	if err := validateImportSourceLinkKey(link.Key); err != nil {
		return err
	}
	if link.ResourceType == "" || strings.TrimSpace(link.ResourceID) == "" || link.JobID.String() == "" || link.CreatedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validateImportSourceLinkKey(key ports.ImportSourceLinkKey) error {
	if key.TenantID.String() == "" || key.InventoryID.String() == "" || key.SourceType == "" || strings.TrimSpace(key.SourceInstanceKey) == "" || key.SourceEntityType == "" || strings.TrimSpace(key.SourceEntityID) == "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validateImportJobResource(record ports.ImportJobResource) error {
	if record.TenantID.String() == "" || record.InventoryID.String() == "" || record.JobID.String() == "" || record.ResourceType == "" || strings.TrimSpace(record.ResourceID) == "" || record.SourceType == "" || strings.TrimSpace(record.SourceInstanceKey) == "" || record.SourceEntityType == "" || strings.TrimSpace(record.SourceEntityID) == "" || record.CreatedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func importLinkPersistenceError(err error) error {
	if strings.Contains(err.Error(), "constraint failed") || strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key value") {
		return ports.ErrConflict
	}
	return err
}
