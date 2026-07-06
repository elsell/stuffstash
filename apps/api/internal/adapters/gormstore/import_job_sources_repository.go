package gormstore

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) ReplaceImportJobSource(ctx context.Context, source ports.ImportJobSourceRecord) error {
	if err := validateImportJobSourceRecord(source); err != nil {
		return err
	}
	model := importJobSourceModelFromRecord(source)
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "job_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"tenant_id", "inventory_id", "key_id", "algorithm", "nonce", "ciphertext", "expires_at", "updated_at"}),
	}).Create(&model).Error
}

func (s Store) ImportJobSource(ctx context.Context, scope ports.ImportJobSourceScope) (ports.ImportJobSourceRecord, bool, error) {
	if err := validateImportJobSourceScope(scope); err != nil {
		return ports.ImportJobSourceRecord{}, false, err
	}
	var model importJobSourceModel
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND inventory_id = ? AND job_id = ?", scope.TenantID.String(), scope.InventoryID.String(), scope.JobID.String()).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ImportJobSourceRecord{}, false, nil
	}
	if err != nil {
		return ports.ImportJobSourceRecord{}, false, err
	}
	record := importJobSourceRecordFromModel(model)
	if err := validateImportJobSourceRecord(record); err != nil {
		return ports.ImportJobSourceRecord{}, false, err
	}
	return record, true, nil
}

func (s Store) DeleteImportJobSource(ctx context.Context, scope ports.ImportJobSourceScope) (bool, error) {
	if err := validateImportJobSourceScope(scope); err != nil {
		return false, err
	}
	result := s.db.WithContext(ctx).
		Where("tenant_id = ? AND inventory_id = ? AND job_id = ?", scope.TenantID.String(), scope.InventoryID.String(), scope.JobID.String()).
		Delete(&importJobSourceModel{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (s Store) DeleteExpiredImportJobSources(ctx context.Context, now time.Time) (int, error) {
	if now.IsZero() {
		return 0, ports.ErrInvalidProviderInput
	}
	result := s.db.WithContext(ctx).Where("expires_at <= ?", now).Delete(&importJobSourceModel{})
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

func (s Store) DeleteVacuumableImportJobSources(ctx context.Context, terminalStatuses []importjob.Status, now time.Time) ([]ports.ImportJobSourceScope, error) {
	if now.IsZero() {
		return nil, ports.ErrInvalidProviderInput
	}
	statuses := make([]string, 0, len(terminalStatuses))
	for _, status := range terminalStatuses {
		if !validImportJobStatus(status) {
			return nil, ports.ErrInvalidProviderInput
		}
		statuses = append(statuses, string(status))
	}
	query := s.db.WithContext(ctx).Where("expires_at <= ?", now)
	if len(statuses) > 0 {
		terminalJobs := s.db.Model(&importJobModel{}).
			Select("id").
			Where("import_jobs.id = import_job_sources.job_id").
			Where("import_jobs.tenant_id = import_job_sources.tenant_id").
			Where("import_jobs.inventory_id = import_job_sources.inventory_id").
			Where("import_jobs.status IN ?", statuses)
		query = query.Or("EXISTS (?)", terminalJobs)
	}
	var models []importJobSourceModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return nil, nil
	}
	scopes := make([]ports.ImportJobSourceScope, 0, len(models))
	jobIDs := make([]string, 0, len(models))
	for _, model := range models {
		scopes = append(scopes, ports.ImportJobSourceScope{
			TenantID:    tenant.ID(model.TenantID),
			InventoryID: inventory.InventoryID(model.InventoryID),
			JobID:       importjob.ID(model.JobID),
		})
		jobIDs = append(jobIDs, model.JobID)
	}
	if err := s.db.WithContext(ctx).Where("job_id IN ?", jobIDs).Delete(&importJobSourceModel{}).Error; err != nil {
		return nil, err
	}
	return scopes, nil
}

func importJobSourceModelFromRecord(source ports.ImportJobSourceRecord) importJobSourceModel {
	return importJobSourceModel{
		JobID:       source.Scope.JobID.String(),
		TenantID:    source.Scope.TenantID.String(),
		InventoryID: source.Scope.InventoryID.String(),
		KeyID:       source.Sealed.KeyID,
		Algorithm:   source.Sealed.Algorithm,
		Nonce:       append([]byte{}, source.Sealed.Nonce...),
		Ciphertext:  append([]byte{}, source.Sealed.Ciphertext...),
		ExpiresAt:   source.ExpiresAt,
		CreatedAt:   source.CreatedAt,
		UpdatedAt:   source.UpdatedAt,
	}
}

func importJobSourceRecordFromModel(model importJobSourceModel) ports.ImportJobSourceRecord {
	return ports.ImportJobSourceRecord{
		Scope: ports.ImportJobSourceScope{
			TenantID:    tenant.ID(model.TenantID),
			InventoryID: inventory.InventoryID(model.InventoryID),
			JobID:       importjob.ID(model.JobID),
		},
		Sealed: ports.SealedImportJobSource{
			KeyID:      model.KeyID,
			Algorithm:  model.Algorithm,
			Nonce:      append([]byte{}, model.Nonce...),
			Ciphertext: append([]byte{}, model.Ciphertext...),
		},
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func validateImportJobSourceRecord(source ports.ImportJobSourceRecord) error {
	if err := validateImportJobSourceScope(source.Scope); err != nil {
		return err
	}
	if source.Sealed.KeyID == "" ||
		source.Sealed.Algorithm != ports.ProviderCredentialAlgorithmAES256GCM ||
		len(source.Sealed.Nonce) != ports.ProviderCredentialAESGCMNonceBytes ||
		len(source.Sealed.Ciphertext) == 0 ||
		source.ExpiresAt.IsZero() ||
		source.CreatedAt.IsZero() ||
		source.UpdatedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validateImportJobSourceScope(scope ports.ImportJobSourceScope) error {
	if scope.TenantID.String() == "" || scope.InventoryID.String() == "" || scope.JobID.String() == "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}
