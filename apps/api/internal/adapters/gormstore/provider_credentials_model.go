package gormstore

import "time"

type providerCredentialModel struct {
	ID                string      `gorm:"primaryKey;size:26"`
	TenantID          string      `gorm:"not null;size:26;index:idx_provider_credentials_one_active,unique,where:superseded_at IS NULL,priority:1;index"`
	Tenant            tenantModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
	ProviderProfileID string      `gorm:"not null;size:64;index:idx_provider_credentials_one_active,unique,where:superseded_at IS NULL,priority:2"`
	Capability        string      `gorm:"not null;size:64;index:idx_provider_credentials_one_active,unique,where:superseded_at IS NULL,priority:3"`
	ProviderKind      string      `gorm:"not null;size:64;index:idx_provider_credentials_one_active,unique,where:superseded_at IS NULL,priority:4"`
	Purpose           string      `gorm:"not null;size:64;index:idx_provider_credentials_one_active,unique,where:superseded_at IS NULL,priority:5"`
	KeyID             string      `gorm:"not null;size:128"`
	Algorithm         string      `gorm:"not null;size:64"`
	Nonce             []byte      `gorm:"not null"`
	Ciphertext        []byte      `gorm:"not null"`
	SupersededAt      *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (providerCredentialModel) TableName() string {
	return "provider_credentials"
}
