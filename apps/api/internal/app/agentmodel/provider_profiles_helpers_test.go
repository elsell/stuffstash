package agentmodel

import (
	"context"
	"testing"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type fakeProviderProfileRepository struct {
	saved           map[string]domain.ProviderProfile
	lastCredential  ports.ProviderCredentialRecord
	lastAuditAction audit.Action
}

func newFakeProviderProfileRepository() *fakeProviderProfileRepository {
	return &fakeProviderProfileRepository{saved: map[string]domain.ProviderProfile{}}
}

func (r *fakeProviderProfileRepository) SaveProviderProfile(_ context.Context, profile domain.ProviderProfile, record audit.Record) error {
	r.saved[profile.ID.String()] = profile
	r.lastAuditAction = record.Action
	return nil
}

func (r *fakeProviderProfileRepository) UpdateProviderProfile(_ context.Context, profile domain.ProviderProfile, record audit.Record) error {
	if _, ok := r.saved[profile.ID.String()]; !ok {
		return ports.ErrForbidden
	}
	r.saved[profile.ID.String()] = profile
	r.lastAuditAction = record.Action
	return nil
}

func (r *fakeProviderProfileRepository) ReplaceProviderProfileCredential(_ context.Context, profile domain.ProviderProfile, credential ports.ProviderCredentialRecord, record audit.Record) error {
	if _, ok := r.saved[profile.ID.String()]; !ok {
		return ports.ErrForbidden
	}
	r.saved[profile.ID.String()] = profile
	r.lastCredential = credential
	r.lastAuditAction = record.Action
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
	return newProviderProfileTestServiceWithCredentials(repository, newFakeProviderCredentialRepository(), &fakeProviderProfileTester{}, authorizer)
}

func newProviderProfileTestServiceWithCredentials(repository *fakeProviderProfileRepository, credentials *fakeProviderCredentialRepository, tester ports.ProviderProfileTester, authorizer ports.Authorizer) Service {
	return New(Dependencies{
		Authorizer:                authorizer,
		ProviderProfiles:          repository,
		ProviderProfileUnitOfWork: repository,
		ProviderCredentialVault:   fakeCredentialVault{repository: credentials, sealer: fakeCredentialSealer{}},
		ProviderProfileTester:     tester,
		IDs:                       fixedIDGenerator{},
		Clock:                     fixedClock{},
	})
}

type fakeProviderCredentialRepository struct {
	saved map[ports.ProviderCredentialScope]ports.ProviderCredentialRecord
}

func newFakeProviderCredentialRepository() *fakeProviderCredentialRepository {
	return &fakeProviderCredentialRepository{saved: map[ports.ProviderCredentialScope]ports.ProviderCredentialRecord{}}
}

func (r *fakeProviderCredentialRepository) ReplaceProviderCredential(_ context.Context, credential ports.ProviderCredentialRecord) error {
	r.saved[credential.Scope] = credential
	return nil
}

func (r *fakeProviderCredentialRepository) ActiveProviderCredential(_ context.Context, scope ports.ProviderCredentialScope) (ports.ProviderCredentialRecord, bool, error) {
	record, ok := r.saved[scope]
	return record, ok, nil
}

func (r *fakeProviderCredentialRepository) ActiveProviderCredentialsExist(context.Context) (bool, error) {
	return len(r.saved) > 0, nil
}

func (r *fakeProviderCredentialRepository) SupersedeActiveProviderCredential(context.Context, ports.ProviderCredentialScope, time.Time) error {
	return nil
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
	return []byte("raw:profile-one"), nil
}

type fakeCredentialVault struct {
	repository *fakeProviderCredentialRepository
	sealer     fakeCredentialSealer
	activeErr  error
}

func (v fakeCredentialVault) PrepareProviderCredential(ctx context.Context, input ports.PrepareProviderCredentialInput) (ports.ProviderCredentialRecord, error) {
	sealed, err := v.sealer.SealProviderCredential(ctx, input.Scope, input.Raw)
	if err != nil {
		return ports.ProviderCredentialRecord{}, err
	}
	return ports.ProviderCredentialRecord{
		ID:        input.ID,
		Scope:     input.Scope,
		Sealed:    sealed,
		CreatedAt: input.CreatedAt,
		UpdatedAt: input.UpdatedAt,
	}, nil
}

func (v fakeCredentialVault) ActiveProviderCredentialMaterial(ctx context.Context, scope ports.ProviderCredentialScope) ([]byte, bool, error) {
	if v.activeErr != nil {
		return nil, false, v.activeErr
	}
	record, found, err := v.repository.ActiveProviderCredential(ctx, scope)
	if err != nil || !found {
		return nil, found, err
	}
	raw, err := v.sealer.UnsealProviderCredential(ctx, scope, record.Sealed)
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

type fakeProviderProfileTester struct {
	lastInput ports.ProviderProfileTestInput
	result    ports.ProviderProfileTestResult
	err       error
}

func (f *fakeProviderProfileTester) TestProviderProfile(_ context.Context, input ports.ProviderProfileTestInput) (ports.ProviderProfileTestResult, error) {
	f.lastInput = input
	if f.err != nil || f.result.Status != "" {
		return f.result, f.err
	}
	return ports.ProviderProfileTestResult{
		ProfileID:    input.Profile.ID.String(),
		Capability:   input.Profile.Capability.String(),
		ProviderKind: input.Profile.ProviderKind.String(),
		Status:       ports.ProviderProfileTestStatusSucceeded,
		Message:      "Provider profile test succeeded.",
		TestedAt:     fixedClock{}.Now(),
	}, nil
}

func testPrincipal() identity.Principal {
	return identity.Principal{ID: identity.PrincipalID("user-one")}
}

func stringPtr(value string) *string {
	return &value
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
