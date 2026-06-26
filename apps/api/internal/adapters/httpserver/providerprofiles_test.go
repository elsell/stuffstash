package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestProviderProfileManagementFlowRedactsCredentials(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newProviderProfileTestApp(t, seededState{
		tenants: []seedTenant{{id: tenantID, name: "Home", owner: "tenant-owner"}},
		ids: []string{
			"01ARZ3NDEKTSV4RRFFQ69G5FAW", "audit-create-profile",
			"credential-one", "audit-replace-credential",
			"audit-enable-profile", "audit-disable-profile", "audit-archive-profile",
		},
	}))

	create := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/provider-profiles", "Bearer dev:tenant-owner", map[string]any{
		"capability":         "language_inference",
		"providerKind":       "gemini",
		"displayName":        "Google Gemini",
		"endpointUrl":        "https://generativelanguage.googleapis.com",
		"modelName":          "gemini-2.5-flash-lite",
		"runtimeOptions":     map[string]any{"temperature": 0.1},
		"capabilityMetadata": map[string]any{"toolCalls": true},
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("expected provider profile create status %d, got %d with body %s", http.StatusCreated, create.Code, create.Body.String())
	}
	created := decodeProviderProfile(t, create)
	if created.Data.ID == "" || created.Data.TenantID != tenantID || created.Data.CredentialStatus != "missing" || created.Data.LifecycleState != "disabled" {
		t.Fatalf("unexpected created provider profile: %+v", created.Data)
	}

	replaceCredential := performRequest(server, http.MethodPut, "/tenants/"+tenantID+"/provider-profiles/"+created.Data.ID+"/credential", "Bearer dev:tenant-owner", map[string]any{
		"purpose":    "api_key",
		"credential": "raw-provider-secret",
	})
	if replaceCredential.Code != http.StatusOK {
		t.Fatalf("expected credential replacement status %d, got %d with body %s", http.StatusOK, replaceCredential.Code, replaceCredential.Body.String())
	}
	if strings.Contains(replaceCredential.Body.String(), "raw-provider-secret") || strings.Contains(replaceCredential.Body.String(), "sealed") || strings.Contains(replaceCredential.Body.String(), "test-key") {
		t.Fatalf("credential replacement leaked secret material: %s", replaceCredential.Body.String())
	}
	withCredential := decodeProviderProfile(t, replaceCredential)
	if withCredential.Data.CredentialStatus != "configured" {
		t.Fatalf("expected configured credential status, got %+v", withCredential.Data)
	}

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/provider-profiles", "Bearer dev:tenant-owner", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected provider profile list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}
	listed := decodeProviderProfileList(t, list)
	if len(listed.Data) != 1 || listed.Data[0].ID != created.Data.ID {
		t.Fatalf("expected listed provider profile, got %+v", listed)
	}

	detail := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/provider-profiles/"+created.Data.ID, "Bearer dev:tenant-owner", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("expected provider profile detail status %d, got %d with body %s", http.StatusOK, detail.Code, detail.Body.String())
	}

	enable := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/provider-profiles/"+created.Data.ID+"/enable", "Bearer dev:tenant-owner", nil)
	if enable.Code != http.StatusOK || decodeProviderProfile(t, enable).Data.LifecycleState != "enabled" {
		t.Fatalf("expected enabled provider profile, got status %d body %s", enable.Code, enable.Body.String())
	}
	disable := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/provider-profiles/"+created.Data.ID+"/disable", "Bearer dev:tenant-owner", nil)
	if disable.Code != http.StatusOK || decodeProviderProfile(t, disable).Data.LifecycleState != "disabled" {
		t.Fatalf("expected disabled provider profile, got status %d body %s", disable.Code, disable.Body.String())
	}
	archive := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/provider-profiles/"+created.Data.ID+"/archive", "Bearer dev:tenant-owner", nil)
	if archive.Code != http.StatusOK || decodeProviderProfile(t, archive).Data.LifecycleState != "archived" {
		t.Fatalf("expected archived provider profile, got status %d body %s", archive.Code, archive.Body.String())
	}
}

func TestProviderProfileEndpointsRejectUnauthorizedUsers(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FB0"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	server := NewServer(":0", newProviderProfileTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Other", owner: "other-owner"},
		},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "tenant-owner"}},
		ids: []string{
			"01ARZ3NDEKTSV4RRFFQ69G5FAW", "audit-create-profile",
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
			"audit-editor-grant", "editor-grant-event", "editor-grant-claim",
		},
	}))

	profileResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/provider-profiles", "Bearer dev:tenant-owner", map[string]any{
		"capability":   "language_inference",
		"providerKind": "gemini",
		"displayName":  "Google Gemini",
	})
	if profileResponse.Code != http.StatusCreated {
		t.Fatalf("expected provider profile create status %d, got %d with body %s", http.StatusCreated, profileResponse.Code, profileResponse.Body.String())
	}
	profileID := decodeProviderProfile(t, profileResponse).Data.ID
	profilePath := "/tenants/" + tenantID + "/provider-profiles/" + profileID

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:tenant-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}
	editorGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:tenant-owner", map[string]string{
		"principalId":  "editor-user",
		"relationship": "editor",
	})
	if editorGrant.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, editorGrant.Code, editorGrant.Body.String())
	}

	credentialBody := map[string]string{"purpose": "api_key", "credential": "blocked-provider-secret"}
	operations := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "create", method: http.MethodPost, path: "/tenants/" + tenantID + "/provider-profiles", body: map[string]any{"capability": "language_inference", "providerKind": "gemini", "displayName": "Blocked"}},
		{name: "list", method: http.MethodGet, path: "/tenants/" + tenantID + "/provider-profiles"},
		{name: "detail", method: http.MethodGet, path: profilePath},
		{name: "replace credential", method: http.MethodPut, path: profilePath + "/credential", body: credentialBody},
		{name: "enable", method: http.MethodPost, path: profilePath + "/enable"},
		{name: "disable", method: http.MethodPost, path: profilePath + "/disable"},
		{name: "archive", method: http.MethodPost, path: profilePath + "/archive"},
	}
	callers := []struct {
		name          string
		authorization string
		status        int
		code          string
		message       string
	}{
		{name: "unauthenticated", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "malformed token", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required", message: "Authentication required."},
		{name: "unrelated user", authorization: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
		{name: "wrong tenant owner", authorization: "Bearer dev:other-owner", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
		{name: "inventory viewer", authorization: "Bearer dev:viewer-user", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
		{name: "inventory editor", authorization: "Bearer dev:editor-user", status: http.StatusForbidden, code: "forbidden", message: "Forbidden."},
	}
	for _, operation := range operations {
		for _, caller := range callers {
			t.Run(operation.name+"/"+caller.name, func(t *testing.T) {
				response := performRequest(server, operation.method, operation.path, caller.authorization, operation.body)
				if response.Code != caller.status {
					t.Fatalf("expected status %d, got %d with body %s", caller.status, response.Code, response.Body.String())
				}
				assertSafeError(t, response, caller.code, caller.message)
				if strings.Contains(response.Body.String(), "blocked-provider-secret") {
					t.Fatalf("credential denial leaked request secret: %s", response.Body.String())
				}
			})
		}
	}
}

func newProviderProfileTestApp(t *testing.T, state seededState) app.App {
	t.Helper()

	ctx := context.Background()
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	seedMemoryStore(t, ctx, store, authorizer, state)

	return app.New(app.Dependencies{
		Observer:                  &fakeObserver{},
		Auth:                      auth.NewLocalDevAuthenticator(),
		Authorizer:                authorizer,
		Tenants:                   store,
		TenantUnitOfWork:          store,
		Inventories:               store,
		InventoryUnitOfWork:       store,
		InventoryAccess:           store,
		InventoryAccessUnitOfWork: store,
		CustomAssetTypes:          store,
		CustomAssetTypeUnitOfWork: store,
		CustomFields:              store,
		CustomFieldUnitOfWork:     store,
		Assets:                    store,
		AssetUnitOfWork:           store,
		Undoables:                 store,
		Search:                    store,
		Attachments:               store,
		AttachmentUnitOfWork:      store,
		Blobs:                     store,
		BlobDeletionOutbox:        store,
		Audit:                     store,
		Outbox:                    store,
		ProviderProfiles:          store,
		ProviderProfileUnitOfWork: store,
		ProviderCredentialSealer:  httpTestCredentialSealer{},
		IDs:                       &fakeIDGenerator{ids: state.ids},
	})
}

type httpTestCredentialSealer struct{}

func (httpTestCredentialSealer) SealProviderCredential(_ context.Context, scope ports.ProviderCredentialScope, raw []byte) (ports.SealedProviderCredential, error) {
	return ports.SealedProviderCredential{
		KeyID:      "test-key",
		Algorithm:  ports.ProviderCredentialAlgorithmAES256GCM,
		Nonce:      []byte("123456789012"),
		Ciphertext: []byte("sealed:" + scope.ProviderProfileID),
	}, nil
}

func (httpTestCredentialSealer) UnsealProviderCredential(context.Context, ports.ProviderCredentialScope, ports.SealedProviderCredential) ([]byte, error) {
	return nil, nil
}

type providerProfileBody struct {
	Data providerProfileResponse `json:"data"`
	Meta responseMeta            `json:"meta"`
}

type providerProfileListBody struct {
	Data []providerProfileResponse `json:"data"`
	Meta responseMeta              `json:"meta"`
}

type providerProfileResponse struct {
	ID                 string         `json:"id"`
	TenantID           string         `json:"tenantId"`
	Capability         string         `json:"capability"`
	ProviderKind       string         `json:"providerKind"`
	DisplayName        string         `json:"displayName"`
	EndpointURL        string         `json:"endpointUrl"`
	ModelName          string         `json:"modelName"`
	RuntimeOptions     map[string]any `json:"runtimeOptions"`
	CapabilityMetadata map[string]any `json:"capabilityMetadata"`
	CredentialStatus   string         `json:"credentialStatus"`
	LifecycleState     string         `json:"lifecycleState"`
}

func decodeProviderProfile(t *testing.T, response *httptest.ResponseRecorder) providerProfileBody {
	t.Helper()

	var body providerProfileBody
	decodeBody(t, response, &body)
	return body
}

func decodeProviderProfileList(t *testing.T, response *httptest.ResponseRecorder) providerProfileListBody {
	t.Helper()

	var body providerProfileListBody
	decodeBody(t, response, &body)
	return body
}
