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
	"gorm.io/gorm/clause"
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
		if err := updateProviderProfileModel(tx, profile, model); err != nil {
			return err
		}
		return createAuditRecord(tx, auditRecord)
	})
}

func (s Store) ReplaceProviderProfileCredential(ctx context.Context, profile agentmodel.ProviderProfile, credential ports.ProviderCredentialRecord, auditRecord audit.Record) error {
	if err := validateProviderCredentialRecord(credential); err != nil {
		return err
	}
	if err := validateProviderCredentialMatchesProfile(profile, credential); err != nil {
		return err
	}
	model := providerProfileModelFromDomain(profile)
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := updateProviderProfileModel(tx, profile, model); err != nil {
			return err
		}
		if err := replaceProviderCredentialInTx(tx, credential, credential.UpdatedAt); err != nil {
			return err
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
	if err := s.db.WithContext(ctx).
		Where(&providerProfileModel{TenantID: tenantID.String()}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).
		Find(&models).Error; err != nil {
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

func updateProviderProfileModel(tx *gorm.DB, profile agentmodel.ProviderProfile, model providerProfileModel) error {
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
			"prompt_template":      model.PromptTemplate,
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
	return nil
}

func validateProviderCredentialMatchesProfile(profile agentmodel.ProviderProfile, credential ports.ProviderCredentialRecord) error {
	if credential.Scope.TenantID.String() != profile.TenantID.String() ||
		credential.Scope.ProviderProfileID != profile.ID.String() ||
		string(credential.Scope.Capability) != profile.Capability.String() ||
		string(credential.Scope.ProviderKind) != profile.ProviderKind.String() {
		return ports.ErrInvalidProviderCredential
	}
	return nil
}
