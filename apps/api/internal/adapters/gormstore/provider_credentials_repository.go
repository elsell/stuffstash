package gormstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s Store) ReplaceProviderCredential(ctx context.Context, credential ports.ProviderCredentialRecord) error {
	if err := validateProviderCredentialRecord(credential); err != nil {
		return err
	}
	now := credential.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceProviderCredentialInTx(tx, credential, now)
	})
}

func (s Store) ActiveProviderCredential(ctx context.Context, scope ports.ProviderCredentialScope) (ports.ProviderCredentialRecord, bool, error) {
	if err := validateProviderCredentialScope(scope); err != nil {
		return ports.ProviderCredentialRecord{}, false, err
	}
	var model providerCredentialModel
	err := providerCredentialScopeQuery(s.db.WithContext(ctx), scope).Where("superseded_at IS NULL").Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ProviderCredentialRecord{}, false, nil
	}
	if err != nil {
		return ports.ProviderCredentialRecord{}, false, err
	}
	return model.toProviderCredentialRecord()
}

func (s Store) ActiveProviderCredentialsExist(ctx context.Context) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&providerCredentialModel{}).Where("superseded_at IS NULL").Limit(1).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s Store) SupersedeActiveProviderCredential(ctx context.Context, scope ports.ProviderCredentialScope, supersededAt time.Time) error {
	if err := validateProviderCredentialScope(scope); err != nil {
		return err
	}
	if supersededAt.IsZero() {
		return ports.ErrInvalidProviderCredential
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return supersedeActiveProviderCredentials(tx, scope, supersededAt)
	})
}

func supersedeActiveProviderCredentials(tx *gorm.DB, scope ports.ProviderCredentialScope, supersededAt time.Time) error {
	return providerCredentialScopeQuery(tx.Model(&providerCredentialModel{}), scope).
		Where("superseded_at IS NULL").
		Updates(map[string]any{
			"superseded_at": supersededAt,
			"updated_at":    supersededAt,
		}).Error
}

func replaceProviderCredentialInTx(tx *gorm.DB, credential ports.ProviderCredentialRecord, now time.Time) error {
	if err := supersedeActiveProviderCredentials(tx, credential.Scope, now); err != nil {
		return err
	}
	model := providerCredentialModelFromRecord(credential, now)
	return tx.Create(&model).Error
}

func providerCredentialModelFromRecord(credential ports.ProviderCredentialRecord, now time.Time) providerCredentialModel {
	return providerCredentialModel{
		ID:                credential.ID,
		TenantID:          credential.Scope.TenantID.String(),
		ProviderProfileID: credential.Scope.ProviderProfileID,
		Capability:        string(credential.Scope.Capability),
		ProviderKind:      string(credential.Scope.ProviderKind),
		Purpose:           string(credential.Scope.Purpose),
		KeyID:             credential.Sealed.KeyID,
		Algorithm:         credential.Sealed.Algorithm,
		Nonce:             append([]byte{}, credential.Sealed.Nonce...),
		Ciphertext:        append([]byte{}, credential.Sealed.Ciphertext...),
		CreatedAt:         timeOrDefault(credential.CreatedAt, now),
		UpdatedAt:         now,
	}
}

func providerCredentialScopeQuery(db *gorm.DB, scope ports.ProviderCredentialScope) *gorm.DB {
	return db.Where(&providerCredentialModel{
		TenantID:          scope.TenantID.String(),
		ProviderProfileID: scope.ProviderProfileID,
		Capability:        string(scope.Capability),
		ProviderKind:      string(scope.ProviderKind),
		Purpose:           string(scope.Purpose),
	})
}

func (m providerCredentialModel) toProviderCredentialRecord() (ports.ProviderCredentialRecord, bool, error) {
	record := ports.ProviderCredentialRecord{
		ID: m.ID,
		Scope: ports.ProviderCredentialScope{
			TenantID:          tenant.ID(m.TenantID),
			ProviderProfileID: m.ProviderProfileID,
			Capability:        ports.ProviderCapability(m.Capability),
			ProviderKind:      ports.ProviderKind(m.ProviderKind),
			Purpose:           ports.ProviderCredentialPurpose(m.Purpose),
		},
		Sealed: ports.SealedProviderCredential{
			KeyID:      m.KeyID,
			Algorithm:  m.Algorithm,
			Nonce:      append([]byte{}, m.Nonce...),
			Ciphertext: append([]byte{}, m.Ciphertext...),
		},
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		SupersededAt: m.SupersededAt,
	}
	if err := validateProviderCredentialRecord(record); err != nil {
		return ports.ProviderCredentialRecord{}, false, fmt.Errorf("invalid provider credential row %q: %w", m.ID, err)
	}
	return record, true, nil
}

func validateProviderCredentialRecord(record ports.ProviderCredentialRecord) error {
	if record.ID == "" ||
		record.Sealed.KeyID == "" ||
		record.Sealed.Algorithm != ports.ProviderCredentialAlgorithmAES256GCM ||
		len(record.Sealed.Nonce) != ports.ProviderCredentialAESGCMNonceBytes ||
		len(record.Sealed.Ciphertext) == 0 {
		return ports.ErrInvalidProviderCredential
	}
	return validateProviderCredentialScope(record.Scope)
}

func validateProviderCredentialScope(scope ports.ProviderCredentialScope) error {
	if scope.TenantID.String() == "" || scope.ProviderProfileID == "" || scope.Capability == "" || scope.ProviderKind == "" || scope.Purpose == "" {
		return ports.ErrInvalidProviderCredential
	}
	switch scope.Capability {
	case ports.ProviderCapabilitySpeechToText, ports.ProviderCapabilityLanguageInference, ports.ProviderCapabilityTextToSpeech:
	default:
		return ports.ErrInvalidProviderCredential
	}
	switch scope.ProviderKind {
	case ports.ProviderKindGemini, ports.ProviderKindOpenAICompatible, ports.ProviderKindLocalHTTP:
	default:
		return ports.ErrInvalidProviderCredential
	}
	switch scope.Purpose {
	case ports.ProviderCredentialPurposeAPIKey, ports.ProviderCredentialPurposeOAuthBearer:
	default:
		return ports.ErrInvalidProviderCredential
	}
	return nil
}

func timeOrDefault(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}
