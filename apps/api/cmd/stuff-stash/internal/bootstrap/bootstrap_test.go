package bootstrap

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/stuffstash/stuff-stash/internal/adapters/dbmigrations"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRunMigrationCommandRejectsMissingAction(t *testing.T) {
	var output bytes.Buffer

	err := RunMigrationCommand(context.Background(), config.Config{DatabaseDSN: "postgres://example"}, nil, &output)
	if err == nil {
		t.Fatalf("expected missing migration action error")
	}
}

func TestRunMigrationCommandRejectsUnknownAction(t *testing.T) {
	var output bytes.Buffer

	err := RunMigrationCommand(context.Background(), config.Config{DatabaseDSN: "postgres://example"}, []string{"sideways"}, &output)
	if err == nil {
		t.Fatalf("expected unknown migration action error")
	}
}

func TestRunMigrationCommandRejectsMissingDSN(t *testing.T) {
	var output bytes.Buffer

	err := RunMigrationCommand(context.Background(), config.Config{}, []string{"up"}, &output)
	if err == nil {
		t.Fatalf("expected missing database dsn error")
	}
}

func TestValidateMigrationStatusRejectsEmptySchema(t *testing.T) {
	err := validateMigrationStatus(dbmigrations.Status{Latest: 3, Empty: true})
	if err == nil {
		t.Fatalf("expected empty schema error")
	}
}

func TestValidateMigrationStatusRejectsDirtySchema(t *testing.T) {
	err := validateMigrationStatus(dbmigrations.Status{Version: 2, Latest: 3, Dirty: true})
	if err == nil {
		t.Fatalf("expected dirty schema error")
	}
}

func TestValidateMigrationStatusRejectsOutdatedSchema(t *testing.T) {
	err := validateMigrationStatus(dbmigrations.Status{Version: 2, Latest: 3})
	if err == nil {
		t.Fatalf("expected outdated schema error")
	}
}

func TestValidateMigrationStatusAcceptsCurrentSchema(t *testing.T) {
	err := validateMigrationStatus(dbmigrations.Status{Version: 3, Latest: 3})
	if err != nil {
		t.Fatalf("validate current schema: %v", err)
	}
}

func TestBuildAuthenticatorAcceptsLocalDevMode(t *testing.T) {
	authenticator, err := buildAuthenticator(context.Background(), config.Config{AuthMode: "local-dev"})
	if err != nil {
		t.Fatalf("build authenticator: %v", err)
	}
	if authenticator == nil {
		t.Fatalf("expected authenticator")
	}
}

func TestBuildAuthenticatorRejectsUnknownMode(t *testing.T) {
	_, err := buildAuthenticator(context.Background(), config.Config{AuthMode: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported mode error")
	}
}

func TestBuildAuthenticatorRejectsIncompleteOIDCConfig(t *testing.T) {
	_, err := buildAuthenticator(context.Background(), config.Config{AuthMode: "oidc"})
	if err == nil {
		t.Fatalf("expected incomplete OIDC config error")
	}
}

func TestBuildAuthorizerAcceptsMemoryMode(t *testing.T) {
	authorizer, closeAuthorizer, err := buildAuthorizer(context.Background(), config.Config{AuthzMode: "memory"})
	if err != nil {
		t.Fatalf("build authorizer: %v", err)
	}
	if authorizer == nil {
		t.Fatalf("expected authorizer")
	}
	if err := closeAuthorizer(); err != nil {
		t.Fatalf("close authorizer: %v", err)
	}
}

func TestBuildAuthorizerRejectsUnknownMode(t *testing.T) {
	_, _, err := buildAuthorizer(context.Background(), config.Config{AuthzMode: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported mode error")
	}
}

func TestBuildRepositoriesAcceptsMemoryMode(t *testing.T) {
	repositories, closeRepositories, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "memory"})
	if err != nil {
		t.Fatalf("build repositories: %v", err)
	}
	if repositories.tenants == nil || repositories.inventories == nil {
		t.Fatalf("expected repositories")
	}
	if err := closeRepositories(); err != nil {
		t.Fatalf("close repositories: %v", err)
	}
}

func TestBuildRepositoriesRejectsUnknownMode(t *testing.T) {
	_, _, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported mode error")
	}
}

func TestBuildRepositoriesRejectsPostgresWithoutDSN(t *testing.T) {
	_, _, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "postgres"})
	if err == nil {
		t.Fatalf("expected missing database dsn error")
	}
}

func TestBuildRepositoriesAcceptsSQLiteMode(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "stuffstash.sqlite")
	repositories, closeRepositories, err := buildRepositories(context.Background(), config.Config{
		RepositoryMode:  "sqlite",
		DatabaseDSN:     dbPath,
		BlobStoragePath: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("build repositories: %v", err)
	}
	if repositories.tenants == nil || repositories.providerCredentials == nil || repositories.blobs == nil {
		t.Fatalf("expected sqlite-backed repositories")
	}
	if err := closeRepositories(); err != nil {
		t.Fatalf("close repositories: %v", err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected sqlite database file: %v", err)
	}
}

func TestBuildRepositoriesSQLitePreservesProviderCredentialSupersession(t *testing.T) {
	ctx := context.Background()
	repositories, closeRepositories, err := buildRepositories(ctx, config.Config{
		RepositoryMode:  "sqlite",
		DatabaseDSN:     filepath.Join(t.TempDir(), "nested", "stuffstash.sqlite"),
		BlobStoragePath: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("build repositories: %v", err)
	}
	defer func() {
		if err := closeRepositories(); err != nil {
			t.Fatalf("close repositories: %v", err)
		}
	}()

	tenantID := tenant.ID("tenant-home")
	tenantName, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	item, ok := tenant.NewTenant(tenantID, tenantName, tenant.LifecycleStateActive)
	if !ok {
		t.Fatalf("expected valid tenant")
	}
	if err := repositories.tenantUnitOfWork.SaveTenant(ctx, item); err != nil {
		t.Fatalf("save tenant: %v", err)
	}

	profile := sqliteBootstrapProviderProfile(t, tenantID)
	if err := repositories.providerProfileUnitOfWork.SaveProviderProfile(ctx, profile, sqliteBootstrapAuditRecord(t, "audit-profile", tenantID, audit.ActionProviderProfileCreated)); err != nil {
		t.Fatalf("save provider profile: %v", err)
	}
	configured, ok := profile.WithCredentialConfigured(profile.UpdatedAt.Add(time.Minute))
	if !ok {
		t.Fatalf("configure provider credential status")
	}
	first := sqliteBootstrapProviderCredential("credential-one", configured, "ciphertext-one", configured.UpdatedAt)
	if err := repositories.providerProfileUnitOfWork.ReplaceProviderProfileCredential(ctx, configured, first, sqliteBootstrapAuditRecord(t, "audit-credential-one", tenantID, audit.ActionProviderProfileCredentialReplaced)); err != nil {
		t.Fatalf("replace first credential: %v", err)
	}
	second := sqliteBootstrapProviderCredential("credential-two", configured, "ciphertext-two", configured.UpdatedAt.Add(time.Minute))
	if err := repositories.providerProfileUnitOfWork.ReplaceProviderProfileCredential(ctx, configured, second, sqliteBootstrapAuditRecord(t, "audit-credential-two", tenantID, audit.ActionProviderProfileCredentialReplaced)); err != nil {
		t.Fatalf("replace second credential: %v", err)
	}

	active, found, err := repositories.providerCredentials.ActiveProviderCredential(ctx, second.Scope)
	if err != nil {
		t.Fatalf("get active credential: %v", err)
	}
	if !found || active.ID != "credential-two" || string(active.Sealed.Ciphertext) != "ciphertext-two" {
		t.Fatalf("unexpected active credential: found=%t credential=%+v", found, active)
	}
	if exists, err := repositories.providerCredentials.ActiveProviderCredentialsExist(ctx); err != nil || !exists {
		t.Fatalf("expected active provider credentials: exists=%t err=%v", exists, err)
	}
}

func TestBuildRepositoriesRejectsSQLiteWithoutDSN(t *testing.T) {
	_, _, err := buildRepositories(context.Background(), config.Config{RepositoryMode: "sqlite"})
	if err == nil {
		t.Fatalf("expected missing database dsn error")
	}
}

func TestReplayLocalDevelopmentAuthorizationRestoresSQLiteOwnerGrants(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "dev.sqlite")
	cfg := config.Config{
		AuthzMode:       "memory",
		RepositoryMode:  "sqlite",
		DatabaseDSN:     dbPath,
		BlobStoragePath: t.TempDir(),
	}
	principal := identity.Principal{ID: identity.PrincipalID("user-one")}
	tenantID := tenant.ID("tenant-home")
	inventoryID := inventory.InventoryID("inventory-home")

	initialRepositories, closeInitialRepositories, err := buildRepositories(ctx, cfg)
	if err != nil {
		t.Fatalf("build initial repositories: %v", err)
	}
	home, ok := tenant.NewTenant(tenantID, tenant.Name("Home"), tenant.LifecycleStateActive)
	if !ok {
		t.Fatalf("expected valid tenant")
	}
	if err := initialRepositories.outbox.SaveTenantAndEnqueueOwnerGrant(ctx, "tenant-owner-event", home, principal, sqliteBootstrapAuditRecord(t, "audit-tenant", tenantID, audit.ActionTenantCreated)); err != nil {
		t.Fatalf("save tenant and enqueue owner grant: %v", err)
	}
	inventoryName, ok := inventory.NewName("Home")
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	homeInventory, ok := inventory.NewInventory(inventoryID, inventory.TenantID(tenantID.String()), inventoryName, inventory.LifecycleStateActive)
	if !ok {
		t.Fatalf("expected valid inventory")
	}
	if err := initialRepositories.outbox.SaveInventoryAndEnqueueOwnerGrant(ctx, "inventory-owner-event", homeInventory, tenantID, principal, sqliteBootstrapAuditRecord(t, "audit-inventory", tenantID, audit.ActionInventoryCreated)); err != nil {
		t.Fatalf("save inventory and enqueue owner grant: %v", err)
	}
	viewerPrincipal := identity.Principal{ID: identity.PrincipalID("viewer-user")}
	viewerGrant := ports.InventoryAccessGrant{
		TenantID:     tenantID,
		InventoryID:  inventoryID,
		PrincipalID:  viewerPrincipal.ID,
		Relationship: ports.InventoryAccessViewer,
	}
	if err := initialRepositories.inventoryAccessUnitOfWork.SaveInventoryAccessGrantAndEnqueue(ctx, "viewer-grant-event", viewerGrant, sqliteBootstrapInventoryAccessAuditRecord(t, "audit-viewer-grant", tenantID, inventoryID, audit.ActionInventoryAccessGranted)); err != nil {
		t.Fatalf("save viewer grant and enqueue authorization event: %v", err)
	}
	editorPrincipal := identity.Principal{ID: identity.PrincipalID("editor-user")}
	editorGrant := ports.InventoryAccessGrant{
		TenantID:     tenantID,
		InventoryID:  inventoryID,
		PrincipalID:  editorPrincipal.ID,
		Relationship: ports.InventoryAccessEditor,
	}
	if err := initialRepositories.inventoryAccessUnitOfWork.SaveInventoryAccessGrantAndEnqueue(ctx, "editor-grant-event", editorGrant, sqliteBootstrapInventoryAccessAuditRecord(t, "audit-editor-grant", tenantID, inventoryID, audit.ActionInventoryAccessGranted)); err != nil {
		t.Fatalf("save editor grant and enqueue authorization event: %v", err)
	}
	viewerRevokeEvent, removed, err := initialRepositories.inventoryAccessUnitOfWork.DeleteInventoryAccessGrantAndClaimRevoke(ctx, "viewer-revoke-event", "viewer-revoke-claim", time.Now().Add(time.Minute), viewerGrant, sqliteBootstrapInventoryAccessAuditRecord(t, "audit-viewer-revoke", tenantID, inventoryID, audit.ActionInventoryAccessRevoked))
	if err != nil {
		t.Fatalf("delete viewer grant and claim revoke: %v", err)
	}
	if !removed {
		t.Fatalf("expected viewer grant to be removed")
	}
	claimedEvents, err := initialRepositories.outbox.ClaimPendingAuthorizationOutboxEvents(ctx, "initial-claim", 10, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim initial authorization events: %v", err)
	}
	if len(claimedEvents) != 4 {
		t.Fatalf("expected four initial authorization events, got %+v", claimedEvents)
	}
	for _, event := range claimedEvents {
		switch event.ID {
		case "editor-grant-event":
			if err := initialRepositories.outbox.MarkAuthorizationOutboxEventDeadLettered(ctx, event.ID, event.ClaimID, "test dead letter"); err != nil {
				t.Fatalf("mark editor authorization event dead-lettered: %v", err)
			}
		default:
			if err := initialRepositories.outbox.MarkAuthorizationOutboxEventProcessed(ctx, event.ID, event.ClaimID); err != nil {
				t.Fatalf("mark initial authorization event processed: %v", err)
			}
		}
	}
	if err := initialRepositories.outbox.MarkAuthorizationOutboxEventProcessed(ctx, viewerRevokeEvent.ID, viewerRevokeEvent.ClaimID); err != nil {
		t.Fatalf("mark viewer revoke authorization event processed: %v", err)
	}
	if err := closeInitialRepositories(); err != nil {
		t.Fatalf("close initial repositories: %v", err)
	}

	reopenedRepositories, closeReopenedRepositories, err := buildRepositories(ctx, cfg)
	if err != nil {
		t.Fatalf("build reopened repositories: %v", err)
	}
	defer func() {
		if err := closeReopenedRepositories(); err != nil {
			t.Fatalf("close reopened repositories: %v", err)
		}
	}()
	authorizer, closeAuthorizer, err := buildAuthorizer(ctx, cfg)
	if err != nil {
		t.Fatalf("build authorizer: %v", err)
	}
	defer func() {
		if err := closeAuthorizer(); err != nil {
			t.Fatalf("close authorizer: %v", err)
		}
	}()
	if err := authorizer.CheckTenant(ctx, principal, ports.TenantPermissionView, tenantID); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected fresh memory authorizer to be empty, got %v", err)
	}

	if err := replayLocalDevelopmentAuthorization(ctx, cfg, authorizer, reopenedRepositories); err != nil {
		t.Fatalf("replay local development authorization: %v", err)
	}
	if err := authorizer.CheckTenant(ctx, principal, ports.TenantPermissionConfigure, tenantID); err != nil {
		t.Fatalf("expected replayed tenant owner access: %v", err)
	}
	if err := authorizer.CheckInventory(ctx, principal, ports.InventoryPermissionConfigure, inventoryID); err != nil {
		t.Fatalf("expected replayed inventory owner access: %v", err)
	}
	if err := authorizer.CheckInventory(ctx, viewerPrincipal, ports.InventoryPermissionView, inventoryID); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected replayed viewer revoke to deny viewer, got %v", err)
	}
	if err := authorizer.CheckInventory(ctx, editorPrincipal, ports.InventoryPermissionEditAsset, inventoryID); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected dead-lettered editor grant to be ignored, got %v", err)
	}
}

func TestReplayLocalDevelopmentAuthorizationSkipsProductionAuthorizationModes(t *testing.T) {
	ctx := context.Background()
	authorizer, closeAuthorizer, err := buildAuthorizer(ctx, config.Config{AuthzMode: "memory"})
	if err != nil {
		t.Fatalf("build authorizer: %v", err)
	}
	defer func() {
		if err := closeAuthorizer(); err != nil {
			t.Fatalf("close authorizer: %v", err)
		}
	}()
	outbox := &replayOnlyOutbox{}

	if err := replayLocalDevelopmentAuthorization(ctx, config.Config{AuthzMode: "spicedb"}, authorizer, repositories{outbox: outbox}); err != nil {
		t.Fatalf("replay should be skipped for non-memory authz: %v", err)
	}
	if outbox.replayCalls != 0 {
		t.Fatalf("expected no replay query for production authz modes, got %d", outbox.replayCalls)
	}
}

func TestBuildBlobStorageAcceptsFilesystemMode(t *testing.T) {
	store, directUploads, err := buildBlobStorage(config.Config{BlobStorageMode: "filesystem", BlobStoragePath: t.TempDir()})
	if err != nil {
		t.Fatalf("build filesystem blob storage: %v", err)
	}
	if store == nil {
		t.Fatalf("expected blob storage")
	}
	if directUploads != nil {
		t.Fatalf("filesystem mode must not expose an unusable direct upload target")
	}
}

func sqliteBootstrapProviderProfile(t *testing.T, tenantID tenant.ID) agentmodel.ProviderProfile {
	t.Helper()

	now := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	profile, ok := agentmodel.NewProviderProfile(agentmodel.ProviderProfileInput{
		ID:                 agentmodel.ProviderProfileID("profile-one"),
		TenantID:           agentmodel.TenantID(tenantID.String()),
		Capability:         agentmodel.ProviderCapabilityLanguageInference,
		ProviderKind:       agentmodel.ProviderKindGemini,
		DisplayName:        agentmodel.DisplayName("Google Gemini"),
		ModelName:          agentmodel.ModelName("gemini-2.5-flash-lite"),
		RuntimeOptionsJSON: []byte(`{"location":"us-central1"}`),
		CapabilityJSON:     []byte(`{"toolCalls":true}`),
		PromptTemplate:     "Answer briefly.",
		CredentialStatus:   agentmodel.CredentialStatusMissing,
		LifecycleState:     agentmodel.ProviderProfileEnabled,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	if !ok {
		t.Fatalf("expected valid provider profile")
	}
	return profile
}

func sqliteBootstrapProviderCredential(id string, profile agentmodel.ProviderProfile, ciphertext string, now time.Time) ports.ProviderCredentialRecord {
	return ports.ProviderCredentialRecord{
		ID: id,
		Scope: ports.ProviderCredentialScope{
			TenantID:          tenant.ID(profile.TenantID.String()),
			ProviderProfileID: profile.ID.String(),
			Capability:        ports.ProviderCapabilityLanguageInference,
			ProviderKind:      ports.ProviderKindGemini,
			Purpose:           ports.ProviderCredentialPurposeAPIKey,
		},
		Sealed: ports.SealedProviderCredential{
			KeyID:      "local-key",
			Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
			Nonce:      []byte("123456789012"),
			Ciphertext: []byte(ciphertext),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func sqliteBootstrapAuditRecord(t *testing.T, id string, tenantID tenant.ID, action audit.Action) audit.Record {
	t.Helper()

	record, ok := audit.NewRecord(
		audit.ID(id),
		audit.TenantID(tenantID.String()),
		"",
		audit.PrincipalID("owner"),
		action,
		audit.SourceAPI,
		audit.TargetProviderProfile,
		"profile-one",
		time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
		"request-1",
		map[string]string{},
	)
	if !ok {
		t.Fatalf("expected valid audit record")
	}
	return record
}

func sqliteBootstrapInventoryAccessAuditRecord(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, action audit.Action) audit.Record {
	t.Helper()

	record, ok := audit.NewRecord(
		audit.ID(id),
		audit.TenantID(tenantID.String()),
		audit.InventoryID(inventoryID.String()),
		audit.PrincipalID("owner"),
		action,
		audit.SourceAPI,
		audit.TargetInventoryAccessGrant,
		id+"-target",
		time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
		"request-1",
		map[string]string{},
	)
	if !ok {
		t.Fatalf("expected valid inventory access audit record")
	}
	return record
}

func TestBuildBlobStorageRejectsUnknownMode(t *testing.T) {
	_, _, err := buildBlobStorage(config.Config{BlobStorageMode: "unknown"})
	if err == nil {
		t.Fatalf("expected unsupported blob storage mode error")
	}
}

func TestBuildBlobStorageRejectsIncompleteS3Config(t *testing.T) {
	_, _, err := buildBlobStorage(config.Config{BlobStorageMode: "s3"})
	if err == nil {
		t.Fatalf("expected incomplete S3 config error")
	}
}

func TestBuildRealtimeVoiceProvidersDisabledByDefault(t *testing.T) {
	stt, lm, tts, err := buildRealtimeVoiceProviders(context.Background(), config.Config{})
	if err != nil {
		t.Fatalf("build providers: %v", err)
	}
	if stt != nil || lm != nil || tts != nil {
		t.Fatalf("expected no realtime voice providers by default")
	}
}

func TestBuildRealtimeVoiceProvidersAcceptsExplicitDevelopmentFakes(t *testing.T) {
	stt, lm, tts, err := buildRealtimeVoiceProviders(context.Background(), config.Config{VoiceDevFakeEnabled: true})
	if err != nil {
		t.Fatalf("build providers: %v", err)
	}
	if stt == nil || lm == nil || tts == nil {
		t.Fatalf("expected development fake realtime voice providers")
	}
}

func TestBuildRealtimeVoiceProvidersRejectsGoogleWithoutProject(t *testing.T) {
	_, _, _, err := buildRealtimeVoiceProviders(context.Background(), config.Config{VoiceGoogleEnabled: true})
	if err == nil || !strings.Contains(err.Error(), "google cloud project") {
		t.Fatalf("expected missing Google project error, got %v", err)
	}
}

func TestBuildRealtimeVoiceProvidersRejectsMalformedGoogleConfig(t *testing.T) {
	_, _, _, err := buildRealtimeVoiceProvidersWithTokenSource(config.Config{
		VoiceGoogleEnabled:    true,
		GoogleCloudProject:    "pianotechpros",
		GoogleCloudLocation:   "us/central1",
		GoogleGeminiModel:     "gemini-test",
		GoogleTTSLanguageCode: "en-US",
		GoogleTTSVoiceName:    "en-US-Neural2-F",
	}, staticBootstrapTokenSource{})
	if err == nil {
		t.Fatalf("expected malformed Google config error")
	}
}

func TestBuildRealtimeVoiceProvidersPrefersGoogleWhenEnabled(t *testing.T) {
	stt, lm, tts, err := buildRealtimeVoiceProvidersWithTokenSource(config.Config{
		VoiceGoogleEnabled:    true,
		VoiceDevFakeEnabled:   true,
		GoogleCloudProject:    "pianotechpros",
		GoogleCloudLocation:   "us-central1",
		GoogleGeminiModel:     "gemini-test",
		GoogleTTSLanguageCode: "en-US",
		GoogleTTSVoiceName:    "en-US-Neural2-F",
	}, staticBootstrapTokenSource{})
	if err != nil {
		t.Fatalf("build providers: %v", err)
	}
	if stt == nil || lm == nil || tts == nil {
		t.Fatalf("expected Google realtime voice providers")
	}
}

func TestBuildRealtimeVoiceProvidersRejectsAccessTokenModeWithoutToken(t *testing.T) {
	_, _, _, err := buildRealtimeVoiceProviders(context.Background(), config.Config{
		VoiceGoogleEnabled:    true,
		GoogleCloudProject:    "pianotechpros",
		GoogleCloudLocation:   "us-central1",
		GoogleGeminiModel:     "gemini-test",
		GoogleTTSLanguageCode: "en-US",
		GoogleTTSVoiceName:    "en-US-Neural2-F",
		GoogleCredentialMode:  config.GoogleCredentialModeAccessToken,
	})
	if err == nil {
		t.Fatalf("expected access token mode without token to fail")
	}
}

func TestBuildRealtimeVoiceProvidersRejectsInvalidGoogleCredentialMode(t *testing.T) {
	_, _, _, err := buildRealtimeVoiceProviders(context.Background(), config.Config{
		VoiceGoogleEnabled:    true,
		GoogleCloudProject:    "pianotechpros",
		GoogleCloudLocation:   "us-central1",
		GoogleGeminiModel:     "gemini-test",
		GoogleTTSLanguageCode: "en-US",
		GoogleTTSVoiceName:    "en-US-Neural2-F",
		GoogleCredentialMode:  "bearer",
		GoogleAccessToken:     "ya29.test",
	})
	if err == nil {
		t.Fatalf("expected invalid Google credential mode to fail")
	}
}

func TestBuildRealtimeVoiceProvidersAcceptsExplicitGoogleAccessTokenMode(t *testing.T) {
	stt, lm, tts, err := buildRealtimeVoiceProviders(context.Background(), config.Config{
		VoiceGoogleEnabled:    true,
		GoogleCloudProject:    "pianotechpros",
		GoogleCloudLocation:   "us-central1",
		GoogleGeminiModel:     "gemini-test",
		GoogleTTSLanguageCode: "en-US",
		GoogleTTSVoiceName:    "en-US-Neural2-F",
		GoogleCredentialMode:  config.GoogleCredentialModeAccessToken,
		GoogleAccessToken:     "ya29.test",
	})
	if err != nil {
		t.Fatalf("build providers: %v", err)
	}
	if stt == nil || lm == nil || tts == nil {
		t.Fatalf("expected Google realtime voice providers")
	}
}

func TestValidateProviderCredentialSealerAllowsNoCredentialsWithoutKey(t *testing.T) {
	repository := fakeProviderCredentialRepository{}

	if err := validateProviderCredentialSealer(context.Background(), config.Config{}, repository); err != nil {
		t.Fatalf("validate provider credential sealer: %v", err)
	}
}

func TestValidateProviderCredentialSealerAcceptsConfiguredKey(t *testing.T) {
	repository := fakeProviderCredentialRepository{}
	cfg := config.Config{
		ProviderCredentialKeyID: "local-key",
		ProviderCredentialKey:   base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)),
	}

	if err := validateProviderCredentialSealer(context.Background(), cfg, repository); err != nil {
		t.Fatalf("validate provider credential sealer: %v", err)
	}
}

func TestValidateProviderCredentialSealerFailsClosedWhenActiveCredentialsNeedMissingKey(t *testing.T) {
	repository := fakeProviderCredentialRepository{activeExists: true}

	if err := validateProviderCredentialSealer(context.Background(), config.Config{}, repository); err == nil {
		t.Fatalf("expected missing provider credential key to fail closed")
	}
}

func TestValidateProviderCredentialSealerRejectsMalformedConfiguredKey(t *testing.T) {
	repository := fakeProviderCredentialRepository{activeExists: true}
	cfg := config.Config{
		ProviderCredentialKeyID: "local-key",
		ProviderCredentialKey:   base64.StdEncoding.EncodeToString([]byte("short")),
	}

	if err := validateProviderCredentialSealer(context.Background(), cfg, repository); err == nil {
		t.Fatalf("expected malformed provider credential key rejection")
	}
}

func TestBootstrapSpiceDBSchemaReadsSchemaFile(t *testing.T) {
	schemaPath := filepath.Join(t.TempDir(), "schema.zed")
	if err := os.WriteFile(schemaPath, []byte("definition user {}"), 0o600); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	bootstrapper := &fakeSchemaBootstrapper{}

	if err := bootstrapSpiceDBSchema(context.Background(), bootstrapper, schemaPath); err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}

	if bootstrapper.schema != "definition user {}" {
		t.Fatalf("expected schema content, got %q", bootstrapper.schema)
	}
}

func TestBootstrapSpiceDBSchemaRetriesTransientFailure(t *testing.T) {
	schemaPath := filepath.Join(t.TempDir(), "schema.zed")
	if err := os.WriteFile(schemaPath, []byte("definition user {}"), 0o600); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	bootstrapper := &fakeSchemaBootstrapper{failuresRemaining: 1}

	if err := bootstrapSpiceDBSchema(context.Background(), bootstrapper, schemaPath); err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}

	if bootstrapper.calls != 2 {
		t.Fatalf("expected 2 bootstrap attempts, got %d", bootstrapper.calls)
	}
}

type fakeSchemaBootstrapper struct {
	failuresRemaining int
	calls             int
	schema            string
}

type fakeProviderCredentialRepository struct {
	activeExists bool
}

type replayOnlyOutbox struct {
	replayCalls int
}

func (o *replayOnlyOutbox) SaveTenantAndEnqueueOwnerGrant(context.Context, string, tenant.Tenant, identity.Principal, audit.Record) error {
	return nil
}

func (o *replayOnlyOutbox) SaveInventoryAndEnqueueOwnerGrant(context.Context, string, inventory.Inventory, tenant.ID, identity.Principal, audit.Record) error {
	return nil
}

func (o *replayOnlyOutbox) ListAuthorizationOutboxReplayEvents(context.Context) ([]ports.AuthorizationOutboxEvent, error) {
	o.replayCalls++
	return nil, nil
}

func (o *replayOnlyOutbox) ClaimAuthorizationOutboxEvent(context.Context, string, string, time.Time) (ports.AuthorizationOutboxEvent, bool, error) {
	return ports.AuthorizationOutboxEvent{}, false, nil
}

func (o *replayOnlyOutbox) ClaimPendingAuthorizationOutboxEvents(context.Context, string, int, time.Time, time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	return nil, nil
}

func (o *replayOnlyOutbox) MarkAuthorizationOutboxEventProcessed(context.Context, string, string) error {
	return nil
}

func (o *replayOnlyOutbox) MarkAuthorizationOutboxEventFailed(context.Context, string, string, string) error {
	return nil
}

func (o *replayOnlyOutbox) MarkAuthorizationOutboxEventDeadLettered(context.Context, string, string, string) error {
	return nil
}

func (f fakeProviderCredentialRepository) ReplaceProviderCredential(context.Context, ports.ProviderCredentialRecord) error {
	return nil
}

func (f fakeProviderCredentialRepository) ActiveProviderCredential(context.Context, ports.ProviderCredentialScope) (ports.ProviderCredentialRecord, bool, error) {
	return ports.ProviderCredentialRecord{}, false, nil
}

func (f fakeProviderCredentialRepository) ActiveProviderCredentialsExist(context.Context) (bool, error) {
	return f.activeExists, nil
}

func (f fakeProviderCredentialRepository) SupersedeActiveProviderCredential(context.Context, ports.ProviderCredentialScope, time.Time) error {
	return nil
}

type staticBootstrapTokenSource struct{}

func (staticBootstrapTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "test-token", TokenType: "Bearer"}, nil
}

func (f *fakeSchemaBootstrapper) BootstrapSchema(_ context.Context, schema string) error {
	f.calls++
	f.schema = schema
	if f.failuresRemaining > 0 {
		f.failuresRemaining--
		return errors.New("not ready")
	}
	return nil
}
