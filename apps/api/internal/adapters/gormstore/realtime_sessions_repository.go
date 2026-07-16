package gormstore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) SaveRealtimeSession(ctx context.Context, record ports.RealtimeSessionRecord) error {
	if err := validateRealtimeSessionRecord(record); err != nil {
		return err
	}
	model := realtimeSessionModelFromRecord(record)
	return s.db.WithContext(ctx).Create(&model).Error
}

func (s Store) UpdateRealtimeSessionOutcome(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, sessionID string, outcome ports.RealtimeSessionOutcome) error {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(sessionID) == "" || validateRealtimeSessionOutcome(outcome) != nil {
		return ports.ErrInvalidProviderInput
	}
	result := s.db.WithContext(ctx).Model(&realtimeSessionModel{}).
		Where(&realtimeSessionModel{
			TenantID:    tenantID.String(),
			InventoryID: inventoryID.String(),
			ID:          sessionID,
			State:       string(ports.RealtimeSessionStateStarted),
		}).
		Where(clause.Lte{Column: clause.Column{Name: "started_at"}, Value: outcome.At}).
		Updates(map[string]any{
			"state":             string(outcome.State),
			"last_activity_at":  outcome.At,
			"ended_at":          outcome.At,
			"safe_failure_code": strings.TrimSpace(outcome.SafeFailureCode),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func (s Store) RealtimeSessionByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, sessionID string) (ports.RealtimeSessionRecord, bool, error) {
	if tenantID.String() == "" || inventoryID.String() == "" || strings.TrimSpace(sessionID) == "" {
		return ports.RealtimeSessionRecord{}, false, ports.ErrInvalidProviderInput
	}
	var model realtimeSessionModel
	err := s.db.WithContext(ctx).Where(&realtimeSessionModel{
		TenantID:    tenantID.String(),
		InventoryID: inventoryID.String(),
		ID:          sessionID,
	}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.RealtimeSessionRecord{}, false, nil
	}
	if err != nil {
		return ports.RealtimeSessionRecord{}, false, err
	}
	return realtimeSessionRecordFromModel(model)
}

func realtimeSessionModelFromRecord(record ports.RealtimeSessionRecord) realtimeSessionModel {
	var endedAt *time.Time
	if !record.EndedAt.IsZero() {
		ended := record.EndedAt
		endedAt = &ended
	}
	return realtimeSessionModel{
		ID:                         record.ID,
		TenantID:                   record.TenantID.String(),
		InventoryID:                record.InventoryID.String(),
		PrincipalID:                record.PrincipalID.String(),
		Source:                     strings.TrimSpace(record.Source),
		State:                      string(record.State),
		SpeechToTextProfileID:      strings.TrimSpace(record.SpeechToTextProfileID),
		LanguageInferenceProfileID: strings.TrimSpace(record.LanguageInferenceProfileID),
		TextToSpeechProfileID:      strings.TrimSpace(record.TextToSpeechProfileID),
		SafeFailureCode:            strings.TrimSpace(record.SafeFailureCode),
		StartedAt:                  record.StartedAt,
		LastActivityAt:             record.LastActivityAt,
		EndedAt:                    endedAt,
	}
}

func realtimeSessionRecordFromModel(model realtimeSessionModel) (ports.RealtimeSessionRecord, bool, error) {
	record := ports.RealtimeSessionRecord{
		ID:                         model.ID,
		TenantID:                   tenant.ID(model.TenantID),
		InventoryID:                inventory.InventoryID(model.InventoryID),
		PrincipalID:                identity.PrincipalID(model.PrincipalID),
		Source:                     model.Source,
		State:                      ports.RealtimeSessionState(model.State),
		SpeechToTextProfileID:      model.SpeechToTextProfileID,
		LanguageInferenceProfileID: model.LanguageInferenceProfileID,
		TextToSpeechProfileID:      model.TextToSpeechProfileID,
		StartedAt:                  model.StartedAt,
		LastActivityAt:             model.LastActivityAt,
		SafeFailureCode:            model.SafeFailureCode,
	}
	if model.EndedAt != nil {
		record.EndedAt = *model.EndedAt
	}
	if err := validateRealtimeSessionReadRecord(record); err != nil {
		return ports.RealtimeSessionRecord{}, false, fmt.Errorf("invalid realtime session row %q: %w", model.ID, err)
	}
	return record, true, nil
}

func validateRealtimeSessionRecord(record ports.RealtimeSessionRecord) error {
	if strings.TrimSpace(record.ID) == "" ||
		record.TenantID.String() == "" ||
		record.InventoryID.String() == "" ||
		record.PrincipalID.String() == "" ||
		strings.TrimSpace(record.Source) == "" ||
		record.State != ports.RealtimeSessionStateStarted ||
		record.StartedAt.IsZero() ||
		record.LastActivityAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if !record.EndedAt.IsZero() || strings.TrimSpace(record.SafeFailureCode) != "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validateRealtimeSessionReadRecord(record ports.RealtimeSessionRecord) error {
	if strings.TrimSpace(record.ID) == "" ||
		record.TenantID.String() == "" ||
		record.InventoryID.String() == "" ||
		record.PrincipalID.String() == "" ||
		strings.TrimSpace(record.Source) == "" ||
		!validRealtimeSessionState(record.State) ||
		record.StartedAt.IsZero() ||
		record.LastActivityAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if record.State == ports.RealtimeSessionStateStarted {
		if !record.EndedAt.IsZero() || strings.TrimSpace(record.SafeFailureCode) != "" {
			return ports.ErrInvalidProviderInput
		}
		return nil
	}
	if record.EndedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if record.State == ports.RealtimeSessionStateFailed && strings.TrimSpace(record.SafeFailureCode) == "" {
		return ports.ErrInvalidProviderInput
	}
	if record.State != ports.RealtimeSessionStateFailed && strings.TrimSpace(record.SafeFailureCode) != "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func validRealtimeSessionState(state ports.RealtimeSessionState) bool {
	switch state {
	case ports.RealtimeSessionStateStarted, ports.RealtimeSessionStateCompleted, ports.RealtimeSessionStateFailed, ports.RealtimeSessionStateCancelled:
		return true
	default:
		return false
	}
}

func validRealtimeSessionFinalState(state ports.RealtimeSessionState) bool {
	switch state {
	case ports.RealtimeSessionStateCompleted, ports.RealtimeSessionStateFailed, ports.RealtimeSessionStateCancelled:
		return true
	default:
		return false
	}
}

func validateRealtimeSessionOutcome(outcome ports.RealtimeSessionOutcome) error {
	if !validRealtimeSessionFinalState(outcome.State) || outcome.At.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	if outcome.State == ports.RealtimeSessionStateFailed && strings.TrimSpace(outcome.SafeFailureCode) == "" {
		return ports.ErrInvalidProviderInput
	}
	if outcome.State != ports.RealtimeSessionStateFailed && strings.TrimSpace(outcome.SafeFailureCode) != "" {
		return ports.ErrInvalidProviderInput
	}
	return nil
}
