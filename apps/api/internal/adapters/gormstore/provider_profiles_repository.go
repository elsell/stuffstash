package gormstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func (s Store) SaveProviderProfile(ctx context.Context, profile agentmodel.ProviderProfile, auditRecord audit.Record) error {
	model := providerProfileModelFromDomain(profile)
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model).Error; err != nil {
			return providerProfileWriteError(err)
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) UpdateProviderProfile(ctx context.Context, profile agentmodel.ProviderProfile, auditRecord audit.Record) error {
	model := providerProfileModelFromDomain(profile)
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&providerProfileModel{}).
			Where(&providerProfileModel{ID: profile.ID.String(), TenantID: profile.TenantID.String()}).
			Updates(map[string]any{
				"capability":           model.Capability,
				"provider_kind":        model.ProviderKind,
				"display_name":         model.DisplayName,
				"endpoint_url":         model.EndpointURL,
				"model_name":           model.ModelName,
				"runtime_options_json": model.RuntimeOptionsJSON,
				"capability_json":      model.CapabilityJSON,
				"credential_status":    model.CredentialStatus,
				"lifecycle_state":      model.LifecycleState,
				"last_tested_at":       model.LastTestedAt,
				"updated_at":           model.UpdatedAt,
			})
		if result.Error != nil {
			return providerProfileWriteError(result.Error)
		}
		if result.RowsAffected == 0 {
			return ports.ErrForbidden
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) ProviderProfileByID(ctx context.Context, tenantID tenant.ID, profileID agentmodel.ProviderProfileID) (agentmodel.ProviderProfile, bool, error) {
	var model providerProfileModel
	err := s.db.WithContext(ctx).Where(&providerProfileModel{ID: profileID.String(), TenantID: tenantID.String()}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return agentmodel.ProviderProfile{}, false, nil
	}
	if err != nil {
		return agentmodel.ProviderProfile{}, false, err
	}
	profile, ok := model.toDomain()
	if !ok {
		return agentmodel.ProviderProfile{}, false, fmt.Errorf("invalid provider profile row %q", model.ID)
	}
	return profile, true, nil
}

func (s Store) ListProviderProfiles(ctx context.Context, tenantID tenant.ID) ([]agentmodel.ProviderProfile, error) {
	var models []providerProfileModel
	if err := s.db.WithContext(ctx).Where(&providerProfileModel{TenantID: tenantID.String()}).Order("created_at ASC, id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	profiles := make([]agentmodel.ProviderProfile, 0, len(models))
	for _, model := range models {
		profile, ok := model.toDomain()
		if !ok {
			return nil, fmt.Errorf("invalid provider profile row %q", model.ID)
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func providerProfileWriteError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "constraint failed") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ports.ErrConflict
	}
	return err
}
