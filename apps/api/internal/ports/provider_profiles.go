package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type ProviderProfileRepository interface {
	ProviderProfileByID(ctx context.Context, tenantID tenant.ID, profileID agentmodel.ProviderProfileID) (agentmodel.ProviderProfile, bool, error)
	ListProviderProfiles(ctx context.Context, tenantID tenant.ID) ([]agentmodel.ProviderProfile, error)
}

type ProviderProfileUnitOfWork interface {
	SaveProviderProfile(ctx context.Context, profile agentmodel.ProviderProfile, auditRecord audit.Record) error
	UpdateProviderProfile(ctx context.Context, profile agentmodel.ProviderProfile, auditRecord audit.Record) error
	ReplaceProviderProfileCredential(ctx context.Context, profile agentmodel.ProviderProfile, credential ProviderCredentialRecord, auditRecord audit.Record) error
}

type ProviderProfileTestStatus string

const (
	ProviderProfileTestStatusSucceeded ProviderProfileTestStatus = "succeeded"
	ProviderProfileTestStatusFailed    ProviderProfileTestStatus = "failed"
)

type ProviderProfileTestInput struct {
	Profile           agentmodel.ProviderProfile
	CredentialPurpose ProviderCredentialPurpose
	Credential        []byte
	TestedAt          time.Time
}

type ProviderProfileTestResult struct {
	ProfileID    string
	Capability   string
	ProviderKind string
	Status       ProviderProfileTestStatus
	Message      string
	TestedAt     time.Time
}

type ProviderProfileTester interface {
	TestProviderProfile(ctx context.Context, input ProviderProfileTestInput) (ProviderProfileTestResult, error)
}
