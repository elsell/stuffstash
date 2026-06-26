package agentmodel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestServiceCreatesTenantProviderProfile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})

	profile, err := service.CreateProviderProfile(ctx, CreateProviderProfileInput{
		Principal:          testPrincipal(),
		Source:             audit.SourceAPI,
		RequestID:          "request-1",
		TenantID:           tenant.ID("tenant-home"),
		Capability:         "language_inference",
		ProviderKind:       "gemini",
		DisplayName:        "Google Gemini",
		ModelName:          "gemini-2.5-flash-lite",
		RuntimeOptionsJSON: []byte(`{"temperature":0.1}`),
		CapabilityJSON:     []byte(`{"toolCalls":true}`),
		Enable:             true,
	})
	if err != nil {
		t.Fatalf("create provider profile: %v", err)
	}
	if profile.TenantID.String() != "tenant-home" || profile.LifecycleState != domain.ProviderProfileEnabled {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if repository.saved[profile.ID.String()].DisplayName.String() != "Google Gemini" {
		t.Fatalf("expected profile to be saved: %+v", repository.saved)
	}
}

func TestServiceRejectsProviderProfileWithoutTenantConfigurePermission(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service := newProviderProfileTestService(newFakeProviderProfileRepository(), denyTenantAuthorizer{})

	_, err := service.CreateProviderProfile(ctx, CreateProviderProfileInput{
		Principal:    testPrincipal(),
		Source:       audit.SourceAPI,
		RequestID:    "request-1",
		TenantID:     tenant.ID("tenant-home"),
		Capability:   "language_inference",
		ProviderKind: "gemini",
		DisplayName:  "Google Gemini",
		ModelName:    "gemini-2.5-flash-lite",
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceRejectsRawCredentialMaterialInProfileInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service := newProviderProfileTestService(newFakeProviderProfileRepository(), allowTenantConfigureAuthorizer{})

	_, err := service.CreateProviderProfile(ctx, CreateProviderProfileInput{
		Principal:          testPrincipal(),
		Source:             audit.SourceAPI,
		RequestID:          "request-1",
		TenantID:           tenant.ID("tenant-home"),
		Capability:         "language_inference",
		ProviderKind:       "gemini",
		DisplayName:        "Google Gemini",
		ModelName:          "gemini-2.5-flash-lite",
		RuntimeOptionsJSON: []byte(`{"apiKey":"secret"}`),
	})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected validation rejection, got %v", err)
	}
}

func TestServiceReplacesProviderProfileCredentialWithSealedMaterial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	repository.saved[profile.ID.String()] = profile

	updated, err := service.ReplaceProviderProfileCredential(ctx, ReplaceProviderProfileCredentialInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
		Purpose:   "api_key",
		Raw:       []byte("raw-provider-secret"),
	})
	if err != nil {
		t.Fatalf("replace provider profile credential: %v", err)
	}
	if updated.CredentialStatus != domain.CredentialStatusConfigured {
		t.Fatalf("expected configured credential status, got %+v", updated)
	}
	if string(repository.lastCredential.Sealed.Ciphertext) == "raw-provider-secret" || len(repository.lastCredential.Sealed.Ciphertext) == 0 {
		t.Fatalf("expected sealed credential material, got %+v", repository.lastCredential)
	}
	if repository.lastCredential.Scope.TenantID.String() != "tenant-home" ||
		repository.lastCredential.Scope.ProviderProfileID != "profile-one" ||
		repository.lastCredential.Scope.Capability != ports.ProviderCapabilityLanguageInference ||
		repository.lastCredential.Scope.ProviderKind != ports.ProviderKindGemini ||
		repository.lastCredential.Scope.Purpose != ports.ProviderCredentialPurposeAPIKey {
		t.Fatalf("unexpected credential scope: %+v", repository.lastCredential.Scope)
	}
}

func TestServiceRejectsCredentialReplacementWhenSealerUnavailable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := New(Dependencies{
		Authorizer:                allowTenantConfigureAuthorizer{},
		ProviderProfiles:          repository,
		ProviderProfileUnitOfWork: repository,
		IDs:                       fixedIDGenerator{},
		Clock:                     fixedClock{},
	})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	repository.saved[profile.ID.String()] = profile

	_, err := service.ReplaceProviderProfileCredential(ctx, ReplaceProviderProfileCredentialInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
		Purpose:   "api_key",
		Raw:       []byte("raw-provider-secret"),
	})
	if !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("expected precondition rejection, got %v", err)
	}
}

func TestServiceListsAndGetsProviderProfilesByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	first := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	second := mustProviderProfile(t, "profile-two", "tenant-home", domain.ProviderCapabilityTextToSpeech, domain.ProviderProfileEnabled)
	otherTenant := mustProviderProfile(t, "profile-three", "tenant-other", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileEnabled)
	repository.saved[first.ID.String()] = first
	repository.saved[second.ID.String()] = second
	repository.saved[otherTenant.ID.String()] = otherTenant

	listed, err := service.ListProviderProfiles(ctx, ListProviderProfilesInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
	})
	if err != nil {
		t.Fatalf("list provider profiles: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected two tenant profiles, got %+v", listed)
	}

	got, err := service.GetProviderProfile(ctx, GetProviderProfileInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: domain.ProviderProfileID("profile-two"),
	})
	if err != nil {
		t.Fatalf("get provider profile: %v", err)
	}
	if got.ID != second.ID {
		t.Fatalf("unexpected profile: %+v", got)
	}
}

func TestServiceLifecycleCommandsRespectArchivedBoundary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := newFakeProviderProfileRepository()
	service := newProviderProfileTestService(repository, allowTenantConfigureAuthorizer{})
	profile := mustProviderProfile(t, "profile-one", "tenant-home", domain.ProviderCapabilityLanguageInference, domain.ProviderProfileDisabled)
	repository.saved[profile.ID.String()] = profile

	enabled, err := service.EnableProviderProfile(ctx, ProviderProfileLifecycleInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-1",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	})
	if err != nil {
		t.Fatalf("enable provider profile: %v", err)
	}
	if enabled.LifecycleState != domain.ProviderProfileEnabled {
		t.Fatalf("expected enabled profile: %+v", enabled)
	}

	archived, err := service.ArchiveProviderProfile(ctx, ProviderProfileLifecycleInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-2",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	})
	if err != nil {
		t.Fatalf("archive provider profile: %v", err)
	}
	if archived.LifecycleState != domain.ProviderProfileArchived {
		t.Fatalf("expected archived profile: %+v", archived)
	}

	if _, err := service.EnableProviderProfile(ctx, ProviderProfileLifecycleInput{
		Principal: testPrincipal(),
		Source:    audit.SourceAPI,
		RequestID: "request-3",
		TenantID:  tenant.ID("tenant-home"),
		ProfileID: profile.ID,
	}); !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("expected archived enable precondition rejection, got %v", err)
	}
}

type fakeProviderProfileRepository struct {
	saved          map[string]domain.ProviderProfile
	lastCredential ports.ProviderCredentialRecord
}

func newFakeProviderProfileRepository() *fakeProviderProfileRepository {
	return &fakeProviderProfileRepository{saved: map[string]domain.ProviderProfile{}}
}

func (r *fakeProviderProfileRepository) SaveProviderProfile(_ context.Context, profile domain.ProviderProfile, _ audit.Record) error {
	r.saved[profile.ID.String()] = profile
	return nil
}

func (r *fakeProviderProfileRepository) UpdateProviderProfile(_ context.Context, profile domain.ProviderProfile, _ audit.Record) error {
	if _, ok := r.saved[profile.ID.String()]; !ok {
		return ports.ErrForbidden
	}
	r.saved[profile.ID.String()] = profile
	return nil
}

func (r *fakeProviderProfileRepository) ReplaceProviderProfileCredential(_ context.Context, profile domain.ProviderProfile, credential ports.ProviderCredentialRecord, _ audit.Record) error {
	if _, ok := r.saved[profile.ID.String()]; !ok {
		return ports.ErrForbidden
	}
	r.saved[profile.ID.String()] = profile
	r.lastCredential = credential
	return nil
}

func (r *fakeProviderProfileRepository) ProviderProfileByID(_ context.Context, tenantID tenant.ID, profileID domain.ProviderProfileID) (domain.ProviderProfile, bool, error) {
	profile, ok := r.saved[profileID.String()]
	if !ok || profile.TenantID.String() != tenantID.String() {
		return domain.ProviderProfile{}, false, nil
	}
	return profile, true, nil
}

func (r *fakeProviderProfileRepository) ListProviderProfiles(_ context.Context, tenantID tenant.ID) ([]domain.ProviderProfile, error) {
	var profiles []domain.ProviderProfile
	for _, profile := range r.saved {
		if profile.TenantID.String() == tenantID.String() {
			profiles = append(profiles, profile)
		}
	}
	return profiles, nil
}

type allowTenantConfigureAuthorizer struct{}

func (allowTenantConfigureAuthorizer) CheckTenant(_ context.Context, _ identity.Principal, permission ports.TenantPermission, _ tenant.ID) error {
	if permission != ports.TenantPermissionConfigure {
		return ports.ErrForbidden
	}
	return nil
}

func (allowTenantConfigureAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return ports.ErrForbidden
}

func (allowTenantConfigureAuthorizer) ListViewableInventoryIDs(context.Context, identity.Principal, tenant.ID, []inventory.InventoryID) ([]inventory.InventoryID, error) {
	return nil, nil
}

func (allowTenantConfigureAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (allowTenantConfigureAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowTenantConfigureAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowTenantConfigureAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowTenantConfigureAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (allowTenantConfigureAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

type denyTenantAuthorizer struct {
	allowTenantConfigureAuthorizer
}

func (denyTenantAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return ports.ErrForbidden
}

type fixedIDGenerator struct{}

func (fixedIDGenerator) NewID() string {
	return "profile-generated"
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
}

func newProviderProfileTestService(repository *fakeProviderProfileRepository, authorizer ports.Authorizer) Service {
	return New(Dependencies{
		Authorizer:                authorizer,
		ProviderProfiles:          repository,
		ProviderProfileUnitOfWork: repository,
		ProviderCredentialSealer:  fakeCredentialSealer{},
		IDs:                       fixedIDGenerator{},
		Clock:                     fixedClock{},
	})
}

type fakeCredentialSealer struct{}

func (fakeCredentialSealer) SealProviderCredential(_ context.Context, scope ports.ProviderCredentialScope, raw []byte) (ports.SealedProviderCredential, error) {
	if len(raw) == 0 {
		return ports.SealedProviderCredential{}, ports.ErrInvalidProviderCredential
	}
	return ports.SealedProviderCredential{
		KeyID:      "test-key",
		Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
		Nonce:      []byte("123456789012"),
		Ciphertext: []byte("sealed:" + scope.ProviderProfileID),
	}, nil
}

func (fakeCredentialSealer) UnsealProviderCredential(context.Context, ports.ProviderCredentialScope, ports.SealedProviderCredential) ([]byte, error) {
	return nil, nil
}

func testPrincipal() identity.Principal {
	return identity.Principal{ID: identity.PrincipalID("user-one")}
}

func mustProviderProfile(t *testing.T, id string, tenantID string, capability domain.ProviderCapability, lifecycle domain.ProviderProfileLifecycleState) domain.ProviderProfile {
	t.Helper()

	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := domain.NewProviderProfile(domain.ProviderProfileInput{
		ID:                 domain.ProviderProfileID(id),
		TenantID:           domain.TenantID(tenantID),
		Capability:         capability,
		ProviderKind:       domain.ProviderKindGemini,
		DisplayName:        domain.DisplayName("Google Gemini"),
		ModelName:          domain.ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{}`),
		CapabilityJSON:     []byte(`{}`),
		CredentialStatus:   domain.CredentialStatusMissing,
		LifecycleState:     lifecycle,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if !ok {
		t.Fatalf("expected test provider profile to validate")
	}
	return profile
}
