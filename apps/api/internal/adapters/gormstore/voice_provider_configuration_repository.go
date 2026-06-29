package gormstore

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

type voiceProviderConfigurationModel struct {
	TenantID                   string      `gorm:"primaryKey;size:26"`
	Tenant                     tenantModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	SpeechToTextProfileID      string      `gorm:"not null;default:'';size:64"`
	LanguageInferenceProfileID string      `gorm:"not null;default:'';size:64"`
	TextToSpeechProfileID      string      `gorm:"not null;default:'';size:64"`
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

func (voiceProviderConfigurationModel) TableName() string {
	return "voice_provider_configurations"
}

func (s Store) VoiceProviderConfiguration(ctx context.Context, tenantID tenant.ID) (ports.VoiceProviderConfigurationRecord, bool, error) {
	var model voiceProviderConfigurationModel
	err := s.db.WithContext(ctx).Where(&voiceProviderConfigurationModel{TenantID: tenantID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.VoiceProviderConfigurationRecord{}, false, nil
	}
	if err != nil {
		return ports.VoiceProviderConfigurationRecord{}, false, err
	}
	return voiceProviderConfigurationRecordFromModel(model), true, nil
}

func (s Store) SaveVoiceProviderConfiguration(ctx context.Context, record ports.VoiceProviderConfigurationRecord, auditRecord audit.Record) error {
	model := voiceProviderConfigurationModelFromRecord(record)
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func voiceProviderConfigurationModelFromRecord(record ports.VoiceProviderConfigurationRecord) voiceProviderConfigurationModel {
	return voiceProviderConfigurationModel{
		TenantID:                   record.TenantID.String(),
		SpeechToTextProfileID:      record.SpeechToTextProfileID,
		LanguageInferenceProfileID: record.LanguageInferenceProfileID,
		TextToSpeechProfileID:      record.TextToSpeechProfileID,
		CreatedAt:                  record.CreatedAt,
		UpdatedAt:                  record.UpdatedAt,
	}
}

func voiceProviderConfigurationRecordFromModel(model voiceProviderConfigurationModel) ports.VoiceProviderConfigurationRecord {
	return ports.VoiceProviderConfigurationRecord{
		TenantID:                   tenant.ID(model.TenantID),
		SpeechToTextProfileID:      model.SpeechToTextProfileID,
		LanguageInferenceProfileID: model.LanguageInferenceProfileID,
		TextToSpeechProfileID:      model.TextToSpeechProfileID,
		CreatedAt:                  model.CreatedAt,
		UpdatedAt:                  model.UpdatedAt,
	}
}
