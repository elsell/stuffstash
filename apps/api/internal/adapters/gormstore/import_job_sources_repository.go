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

const importJobSourceScopeChunkSize = 100

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
		Where(importJobSourceScopeModel(scope)).
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
		Where(importJobSourceScopeModel(scope)).
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
	result := s.db.WithContext(ctx).Where(clause.Lte{Column: clause.Column{Name: "expires_at"}, Value: now}).Delete(&importJobSourceModel{})
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
	var models []importJobSourceModel
	if err := s.db.WithContext(ctx).Where(clause.Lte{Column: clause.Column{Name: "expires_at"}, Value: now}).Find(&models).Error; err != nil {
		return nil, err
	}
	if len(statuses) > 0 {
		terminalScopes, err := s.terminalImportJobSourceScopes(ctx, statuses)
		if err != nil {
			return nil, err
		}
		for _, chunk := range chunkImportJobSourceScopes(terminalScopes) {
			var terminalModels []importJobSourceModel
			if err := s.db.WithContext(ctx).Where(importJobSourceScopesCondition(chunk)).Find(&terminalModels).Error; err != nil {
				return nil, err
			}
			models = append(models, terminalModels...)
		}
	}
	if len(models) == 0 {
		return nil, nil
	}
	scopes := uniqueImportJobSourceScopes(models)
	if len(scopes) == 0 {
		return nil, nil
	}
	for _, chunk := range chunkImportJobSourceScopes(scopes) {
		deleteQuery := scopedImportJobSourceDeleteQuery(s.db.WithContext(ctx), chunk)
		if err := deleteQuery.Delete(&importJobSourceModel{}).Error; err != nil {
			return nil, err
		}
	}
	return scopes, nil
}

func uniqueImportJobSourceScopes(models []importJobSourceModel) []ports.ImportJobSourceScope {
	scopes := make([]ports.ImportJobSourceScope, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		scope := ports.ImportJobSourceScope{
			TenantID:    tenant.ID(model.TenantID),
			InventoryID: inventory.InventoryID(model.InventoryID),
			JobID:       importjob.ID(model.JobID),
		}
		key := importJobSourceScopeKey(scope)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		scopes = append(scopes, scope)
	}
	return scopes
}

func scopedImportJobSourceDeleteQuery(db *gorm.DB, scopes []ports.ImportJobSourceScope) *gorm.DB {
	return db.Where(importJobSourceScopesCondition(scopes))
}

func (s Store) terminalImportJobSourceScopes(ctx context.Context, statuses []string) ([]ports.ImportJobSourceScope, error) {
	var jobs []importJobModel
	if err := s.db.WithContext(ctx).
		Select("id", "tenant_id", "inventory_id").
		Where(clause.IN{Column: clause.Column{Name: "status"}, Values: stringValues(statuses)}).
		Find(&jobs).Error; err != nil {
		return nil, err
	}
	scopes := make([]ports.ImportJobSourceScope, 0, len(jobs))
	for _, job := range jobs {
		scopes = append(scopes, ports.ImportJobSourceScope{
			TenantID:    tenant.ID(job.TenantID),
			InventoryID: inventory.InventoryID(job.InventoryID),
			JobID:       importjob.ID(job.ID),
		})
	}
	return scopes, nil
}

func chunkImportJobSourceScopes(scopes []ports.ImportJobSourceScope) [][]ports.ImportJobSourceScope {
	if len(scopes) == 0 {
		return nil
	}
	chunks := make([][]ports.ImportJobSourceScope, 0, (len(scopes)+importJobSourceScopeChunkSize-1)/importJobSourceScopeChunkSize)
	for start := 0; start < len(scopes); start += importJobSourceScopeChunkSize {
		end := start + importJobSourceScopeChunkSize
		if end > len(scopes) {
			end = len(scopes)
		}
		chunks = append(chunks, scopes[start:end])
	}
	return chunks
}

func importJobSourceScopeKey(scope ports.ImportJobSourceScope) string {
	return scope.TenantID.String() + "\x00" + scope.InventoryID.String() + "\x00" + scope.JobID.String()
}

func importJobSourceScopeModel(scope ports.ImportJobSourceScope) *importJobSourceModel {
	return &importJobSourceModel{
		TenantID:    scope.TenantID.String(),
		InventoryID: scope.InventoryID.String(),
		JobID:       scope.JobID.String(),
	}
}

func importJobSourceScopeCondition(scope ports.ImportJobSourceScope) clause.Expression {
	return clause.And(
		clause.Eq{Column: clause.Column{Name: "tenant_id"}, Value: scope.TenantID.String()},
		clause.Eq{Column: clause.Column{Name: "inventory_id"}, Value: scope.InventoryID.String()},
		clause.Eq{Column: clause.Column{Name: "job_id"}, Value: scope.JobID.String()},
	)
}

func importJobSourceScopesCondition(scopes []ports.ImportJobSourceScope) clause.Expression {
	expressions := make([]clause.Expression, 0, len(scopes))
	for _, scope := range scopes {
		expressions = append(expressions, importJobSourceScopeCondition(scope))
	}
	return clause.Or(expressions...)
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
