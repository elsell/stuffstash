package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveImportJob(ctx context.Context, job importjob.Record) error {
	model, err := importJobModelFromRecord(job)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(&model).Error
}

func (s Store) ImportJobByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" {
		return importjob.Record{}, false, ports.ErrInvalidProviderInput
	}
	var model importJobModel
	err := activeImportJobQuery(s.db.WithContext(ctx), importJobModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		ID:          jobID.String(),
	}).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return importjob.Record{}, false, nil
	}
	if err != nil {
		return importjob.Record{}, false, err
	}
	record, err := importJobRecordFromModel(model)
	if err != nil {
		return importjob.Record{}, false, err
	}
	return record, true, nil
}

func (s Store) ListImportJobs(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.ImportJobPageRequest) ([]importjob.Record, error) {
	if tenantID.String() == "" || inventoryID.String() == "" {
		return nil, ports.ErrInvalidProviderInput
	}
	limit := page.Limit
	if limit <= 0 {
		limit = 50
	}
	var models []importJobModel
	if err := activeImportJobQuery(s.db.WithContext(ctx), importJobModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
	}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}
	jobs := make([]importjob.Record, 0, len(models))
	for _, model := range models {
		job, err := importJobRecordFromModel(model)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s Store) ListImportJobsByStatus(ctx context.Context, page ports.ImportJobStatusPageRequest) ([]importjob.Record, error) {
	if !validImportJobStatus(page.Status) {
		return nil, ports.ErrInvalidProviderInput
	}
	limit := page.Limit
	if limit <= 0 {
		limit = 50
	}
	var models []importJobModel
	if err := activeImportJobQuery(s.db.WithContext(ctx), importJobModel{Status: string(page.Status)}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "updated_at"}}).
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, err
	}
	jobs := make([]importjob.Record, 0, len(models))
	for _, model := range models {
		job, err := importJobRecordFromModel(model)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s Store) UpdateImportJob(ctx context.Context, job importjob.Record) error {
	model, err := importJobModelFromRecord(job)
	if err != nil {
		return err
	}
	result := activeImportJobQuery(s.db.WithContext(ctx).Model(&importJobModel{}), importJobModel{
		TenantID:    job.TenantID.String(),
		InventoryID: job.InventoryID.String(),
		ID:          job.ID.String(),
	}).
		Updates(importJobUpdateMap(model))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s Store) MarkImportJobHistoryRemoved(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, removedAt time.Time, expectedUpdatedAt time.Time) (bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" || removedAt.IsZero() || expectedUpdatedAt.IsZero() {
		return false, ports.ErrInvalidProviderInput
	}
	result := activeImportJobQuery(s.db.WithContext(ctx).Model(&importJobModel{}), importJobModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		ID:          jobID.String(),
		UpdatedAt:   expectedUpdatedAt.UTC(),
	}).
		Updates(map[string]any{
			"history_removed_at": removedAt.UTC(),
			"updated_at":         removedAt.UTC(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func (s Store) UpdateImportJobIfStatus(ctx context.Context, job importjob.Record, expected importjob.Status) (bool, error) {
	if !validImportJobStatus(expected) {
		return false, ports.ErrInvalidProviderInput
	}
	model, err := importJobModelFromRecord(job)
	if err != nil {
		return false, err
	}
	result := activeImportJobQuery(s.db.WithContext(ctx).Model(&importJobModel{}), importJobModel{
		TenantID:    job.TenantID.String(),
		InventoryID: job.InventoryID.String(),
		ID:          job.ID.String(),
		Status:      string(expected),
	}).
		Updates(importJobUpdateMap(model))
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func (s Store) UpdateImportJobProgress(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, progress importjob.Progress, expectedUpdatedAt time.Time) (bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || jobID.String() == "" || progress.Phase == "" || expectedUpdatedAt.IsZero() {
		return false, ports.ErrInvalidProviderInput
	}
	updatedAt := databaseTimestamp(progress.UpdatedAt)
	if updatedAt.IsZero() {
		return false, ports.ErrInvalidProviderInput
	}
	expectedUpdatedAt = databaseTimestamp(expectedUpdatedAt)
	if !updatedAt.After(expectedUpdatedAt) {
		return false, ports.ErrInvalidProviderInput
	}
	var updated bool
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model importJobModel
		err := activeImportJobQuery(tx, importJobModel{
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
			ID:          jobID.String(),
			UpdatedAt:   expectedUpdatedAt.UTC(),
		}).
			First(&model).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		history := []importjob.Progress{}
		if len(model.ProgressHistoryJSON) > 0 {
			if err := json.Unmarshal(model.ProgressHistoryJSON, &history); err != nil {
				return err
			}
		}
		if len(history) == 0 {
			history = importjob.AppendProgressHistory(nil, importjob.Progress{
				Phase:     importjob.Phase(model.ProgressPhase),
				Done:      model.ProgressDone,
				Total:     model.ProgressTotal,
				Message:   model.ProgressMessage,
				UpdatedAt: timeValue(model.ProgressUpdatedAt),
			})
		}
		history = importjob.AppendProgressHistory(history, progress)
		historyJSON, err := json.Marshal(history)
		if err != nil {
			return err
		}
		result := activeImportJobQuery(tx.Model(&importJobModel{}), importJobModel{
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
			ID:          jobID.String(),
			UpdatedAt:   expectedUpdatedAt.UTC(),
		}).
			Updates(map[string]any{
				"progress_phase":        string(progress.Phase),
				"progress_done":         progress.Done,
				"progress_total":        progress.Total,
				"progress_message":      strings.TrimSpace(progress.Message),
				"progress_updated_at":   updatedAt,
				"progress_history_json": historyJSON,
				"updated_at":            updatedAt,
			})
		if result.Error != nil {
			return result.Error
		}
		updated = result.RowsAffected == 1
		return nil
	})
	if err != nil {
		return false, err
	}
	return updated, nil
}

func (s Store) ClaimImportJob(ctx context.Context, job importjob.Record, expectedUpdatedAt time.Time) (bool, error) {
	if expectedUpdatedAt.IsZero() {
		return false, ports.ErrInvalidProviderInput
	}
	expectedUpdatedAt = databaseTimestamp(expectedUpdatedAt)
	model, err := importJobModelFromRecord(job)
	if err != nil {
		return false, err
	}
	result := activeImportJobQuery(s.db.WithContext(ctx).Model(&importJobModel{}), importJobModel{
		TenantID:    job.TenantID.String(),
		InventoryID: job.InventoryID.String(),
		ID:          job.ID.String(),
		UpdatedAt:   expectedUpdatedAt.UTC(),
	}).
		Updates(importJobUpdateMap(model))
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func activeImportJobQuery(db *gorm.DB, scope importJobModel) *gorm.DB {
	return db.Where(&scope).Where(clause.Eq{
		Column: clause.Column{Name: "history_removed_at"},
		Value:  nil,
	})
}

func importJobModelFromRecord(job importjob.Record) (importJobModel, error) {
	if err := validateImportJobRecord(job); err != nil {
		return importJobModel{}, err
	}
	messages, err := json.Marshal(job.Messages)
	if err != nil {
		return importJobModel{}, err
	}
	preview, err := json.Marshal(job.Preview)
	if err != nil {
		return importJobModel{}, err
	}
	progressHistory := job.ProgressHistory
	if len(progressHistory) == 0 {
		progressHistory = importjob.AppendProgressHistory(nil, job.Progress)
	}
	progressHistoryJSON, err := json.Marshal(progressHistory)
	if err != nil {
		return importJobModel{}, err
	}
	return importJobModel{
		ID:                        job.ID.String(),
		TenantID:                  job.TenantID.String(),
		InventoryID:               job.InventoryID.String(),
		ActorID:                   job.ActorID.String(),
		Status:                    string(job.Status),
		SourceType:                string(job.Source.Type),
		SourceName:                strings.TrimSpace(job.Source.Name),
		SourceBaseURL:             strings.TrimSpace(job.Source.BaseURL),
		SourceVersion:             strings.TrimSpace(job.Source.Version),
		SourceImageImport:         strings.TrimSpace(job.Source.ImageImport),
		SourceAllowPrivateNetwork: job.Source.AllowPrivateNetwork,
		SourceAllowInsecureTLS:    job.Source.AllowInsecureTLS,
		SourceFingerprint:         strings.TrimSpace(job.Source.Fingerprint),
		Fields:                    job.Counts.Fields,
		Tags:                      job.Counts.Tags,
		Locations:                 job.Counts.Locations,
		Assets:                    job.Counts.Assets,
		Attachments:               job.Counts.Attachments,
		Warnings:                  job.Counts.Warnings,
		Errors:                    job.Counts.Errors,
		FieldsCreated:             job.Counts.FieldsCreated,
		FieldsExisting:            job.Counts.FieldsExisting,
		TagsCreated:               job.Counts.TagsCreated,
		TagsExisting:              job.Counts.TagsExisting,
		LocationsCreated:          job.Counts.LocationsCreated,
		AssetsCreated:             job.Counts.AssetsCreated,
		AssetsSkipped:             job.Counts.AssetsSkipped,
		AttachmentsCreated:        job.Counts.AttachmentsCreated,
		AttachmentsSkipped:        job.Counts.AttachmentsSkipped,
		RecordsDiscarded:          job.Counts.RecordsDiscarded,
		SourceLinksDiscarded:      job.Counts.SourceLinksDiscarded,
		PreviewJSON:               preview,
		ProgressPhase:             string(job.Progress.Phase),
		ProgressDone:              job.Progress.Done,
		ProgressTotal:             job.Progress.Total,
		ProgressMessage:           strings.TrimSpace(job.Progress.Message),
		ProgressUpdatedAt:         timePointer(job.Progress.UpdatedAt),
		ProgressHistoryJSON:       progressHistoryJSON,
		CancellationMode:          string(job.CancellationMode),
		CancellationRequestID:     strings.TrimSpace(job.CancellationRequestID),
		MessagesJSON:              messages,
		HistoryRemovedAt:          timePointer(job.HistoryRemovedAt),
		StartedAt:                 timePointer(job.StartedAt),
		CompletedAt:               timePointer(job.CompletedAt),
		CreatedAt:                 databaseTimestamp(job.CreatedAt),
		UpdatedAt:                 databaseTimestamp(job.UpdatedAt),
	}, nil
}

func importJobRecordFromModel(model importJobModel) (importjob.Record, error) {
	var messages []importjob.Message
	if err := json.Unmarshal(model.MessagesJSON, &messages); err != nil {
		return importjob.Record{}, err
	}
	var preview importjob.PreviewSummary
	if len(model.PreviewJSON) > 0 {
		if err := json.Unmarshal(model.PreviewJSON, &preview); err != nil {
			return importjob.Record{}, err
		}
	}
	progress := importjob.Progress{
		Phase:     importjob.Phase(model.ProgressPhase),
		Done:      model.ProgressDone,
		Total:     model.ProgressTotal,
		Message:   model.ProgressMessage,
		UpdatedAt: timeValue(model.ProgressUpdatedAt),
	}
	progressHistory := []importjob.Progress{}
	if len(model.ProgressHistoryJSON) > 0 {
		if err := json.Unmarshal(model.ProgressHistoryJSON, &progressHistory); err != nil {
			return importjob.Record{}, err
		}
	}
	if len(progressHistory) == 0 {
		progressHistory = importjob.AppendProgressHistory(nil, progress)
	}
	job := importjob.Record{
		ID:          importjob.ID(model.ID),
		TenantID:    importjob.TenantID(model.TenantID),
		InventoryID: importjob.InventoryID(model.InventoryID),
		ActorID:     importjob.PrincipalID(model.ActorID),
		Status:      importjob.Status(model.Status),
		Source: importjob.SourceRef{
			Type:                importjob.SourceType(model.SourceType),
			Name:                model.SourceName,
			BaseURL:             model.SourceBaseURL,
			Version:             model.SourceVersion,
			ImageImport:         model.SourceImageImport,
			AllowPrivateNetwork: model.SourceAllowPrivateNetwork,
			AllowInsecureTLS:    model.SourceAllowInsecureTLS,
			Fingerprint:         model.SourceFingerprint,
		},
		Counts: importjob.Counts{
			Fields:               model.Fields,
			Tags:                 model.Tags,
			Locations:            model.Locations,
			Assets:               model.Assets,
			Attachments:          model.Attachments,
			Warnings:             model.Warnings,
			Errors:               model.Errors,
			FieldsCreated:        model.FieldsCreated,
			FieldsExisting:       model.FieldsExisting,
			TagsCreated:          model.TagsCreated,
			TagsExisting:         model.TagsExisting,
			LocationsCreated:     model.LocationsCreated,
			AssetsCreated:        model.AssetsCreated,
			AssetsSkipped:        model.AssetsSkipped,
			AttachmentsCreated:   model.AttachmentsCreated,
			AttachmentsSkipped:   model.AttachmentsSkipped,
			RecordsDiscarded:     model.RecordsDiscarded,
			SourceLinksDiscarded: model.SourceLinksDiscarded,
		},
		Preview:               preview,
		Progress:              progress,
		ProgressHistory:       progressHistory,
		CancellationMode:      importjob.CancellationMode(model.CancellationMode),
		CancellationRequestID: model.CancellationRequestID,
		Messages:              messages,
		HistoryRemovedAt:      timeValue(model.HistoryRemovedAt),
		StartedAt:             timeValue(model.StartedAt),
		CompletedAt:           timeValue(model.CompletedAt),
		CreatedAt:             model.CreatedAt,
		UpdatedAt:             model.UpdatedAt,
	}
	if err := validateImportJobRecord(job); err != nil {
		return importjob.Record{}, fmt.Errorf("invalid import job row %q: %w", model.ID, err)
	}
	return job, nil
}

func importJobUpdateMap(model importJobModel) map[string]any {
	return map[string]any{
		"status":                       model.Status,
		"fields":                       model.Fields,
		"tags":                         model.Tags,
		"locations":                    model.Locations,
		"assets":                       model.Assets,
		"attachments":                  model.Attachments,
		"warnings":                     model.Warnings,
		"errors":                       model.Errors,
		"fields_created":               model.FieldsCreated,
		"fields_existing":              model.FieldsExisting,
		"tags_created":                 model.TagsCreated,
		"tags_existing":                model.TagsExisting,
		"locations_created":            model.LocationsCreated,
		"assets_created":               model.AssetsCreated,
		"assets_skipped":               model.AssetsSkipped,
		"attachments_created":          model.AttachmentsCreated,
		"attachments_skipped":          model.AttachmentsSkipped,
		"records_discarded":            model.RecordsDiscarded,
		"source_links_discarded":       model.SourceLinksDiscarded,
		"source_allow_private_network": model.SourceAllowPrivateNetwork,
		"source_allow_insecure_tls":    model.SourceAllowInsecureTLS,
		"preview_json":                 model.PreviewJSON,
		"progress_phase":               model.ProgressPhase,
		"progress_done":                model.ProgressDone,
		"progress_total":               model.ProgressTotal,
		"progress_message":             model.ProgressMessage,
		"progress_updated_at":          model.ProgressUpdatedAt,
		"progress_history_json":        model.ProgressHistoryJSON,
		"cancellation_mode":            model.CancellationMode,
		"cancellation_request_id":      model.CancellationRequestID,
		"messages_json":                model.MessagesJSON,
		"started_at":                   model.StartedAt,
		"completed_at":                 model.CompletedAt,
		"updated_at":                   model.UpdatedAt,
	}
}

func validateImportJobRecord(job importjob.Record) error {
	if job.ID.String() == "" ||
		job.TenantID.String() == "" ||
		job.InventoryID.String() == "" ||
		job.ActorID.String() == "" ||
		!validImportJobStatus(job.Status) ||
		job.Source.Type == "" ||
		strings.TrimSpace(job.Source.Name) == "" ||
		strings.TrimSpace(job.Source.Fingerprint) == "" ||
		job.CreatedAt.IsZero() ||
		job.UpdatedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if job.Progress.Phase == "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validImportJobStatus(status importjob.Status) bool {
	switch status {
	case importjob.StatusPreviewed, importjob.StatusRunning, importjob.StatusSucceeded, importjob.StatusFailed, importjob.StatusCancelRequested, importjob.StatusCancelledKept, importjob.StatusCancelledDiscarded, importjob.StatusDiscardFailed:
		return true
	default:
		return false
	}
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	copy := databaseTimestamp(value)
	return &copy
}

func timeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return databaseTimestamp(*value)
}

func databaseTimestamp(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC().Truncate(time.Microsecond)
}
