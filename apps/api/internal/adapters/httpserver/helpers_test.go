package httpserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestApp(observer ports.Observer, ids ...string) app.App {
	return newTestAppWithAuthorizer(observer, memory.NewAuthorizer(), ids...)
}

func newTestAppWithAuthorizer(observer ports.Observer, authorizer ports.Authorizer, ids ...string) app.App {
	store := memory.NewStore()
	return app.New(app.Dependencies{
		Observer:         observer,
		Auth:             auth.NewLocalDevAuthenticator(),
		Authorizer:       authorizer,
		Tenants:          store,
		Inventories:      store,
		CustomAssetTypes: store,
		CustomFields:     store,
		Assets:           store,
		Search:           store,
		Attachments:      store,
		Blobs:            store,
		Audit:            store,
		Outbox:           store,
		IDs:              &fakeIDGenerator{ids: ids},
	})
}

func newSeededTestApp(t *testing.T, state seededState) app.App {
	return newSeededTestAppWithBlob(t, state, nil)
}

func newSeededTestAppWithBlob(t *testing.T, state seededState, blobStorage ports.BlobStorage) app.App {
	return newSeededTestAppWithBlobAndAuthorizer(t, state, blobStorage, memory.NewAuthorizer())
}

func newSeededTestAppWithAuthorizer(t *testing.T, state seededState, authorizer ports.Authorizer) app.App {
	return newSeededTestAppWithBlobAndAuthorizer(t, state, nil, authorizer)
}

func newSeededTestAppWithBlobAndAuthorizer(t *testing.T, state seededState, blobStorage ports.BlobStorage, authorizer ports.Authorizer) app.App {
	t.Helper()

	ctx := context.Background()
	store := memory.NewStore()
	seedMemoryStore(t, ctx, store, authorizer, state)

	if blobStorage == nil {
		blobStorage = store
	}

	return app.New(app.Dependencies{
		Observer:         &fakeObserver{},
		Auth:             auth.NewLocalDevAuthenticator(),
		Authorizer:       authorizer,
		Tenants:          store,
		Inventories:      store,
		CustomAssetTypes: store,
		CustomFields:     store,
		Assets:           store,
		Search:           store,
		Attachments:      store,
		Blobs:            blobStorage,
		Audit:            store,
		Outbox:           store,
		IDs:              &fakeIDGenerator{ids: state.ids},
		InvitationTTL:    state.invitationTTL,
	})
}

func newSeededTestAppWithStoreAndAuthorizer(t *testing.T, state seededState, store *memory.Store, authorizer ports.Authorizer) app.App {
	t.Helper()

	seedMemoryStore(t, context.Background(), store, authorizer, state)

	return app.New(app.Dependencies{
		Observer:         &fakeObserver{},
		Auth:             auth.NewLocalDevAuthenticator(),
		Authorizer:       authorizer,
		Tenants:          store,
		Inventories:      store,
		CustomAssetTypes: store,
		CustomFields:     store,
		Assets:           store,
		Search:           store,
		Attachments:      store,
		Blobs:            store,
		Audit:            store,
		Outbox:           store,
		IDs:              &fakeIDGenerator{ids: state.ids},
	})
}

func seedMemoryStore(t *testing.T, ctx context.Context, store *memory.Store, authorizer ports.Authorizer, state seededState) {
	t.Helper()

	for _, item := range state.tenants {
		tenantID := tenant.ID(item.id)
		name, ok := tenant.NewName(item.name)
		if !ok {
			t.Fatalf("invalid tenant name %q", item.name)
		}
		if err := store.SaveTenant(ctx, tenant.Tenant{ID: tenantID, Name: name}); err != nil {
			t.Fatalf("save tenant: %v", err)
		}
		if item.owner != "" {
			if err := authorizer.GrantTenantOwner(ctx, principal(item.owner), tenantID); err != nil {
				t.Fatalf("grant tenant owner: %v", err)
			}
		}
	}

	for _, item := range state.inventories {
		name, ok := inventory.NewName(item.name)
		if !ok {
			t.Fatalf("invalid inventory name %q", item.name)
		}
		inventoryID := inventory.InventoryID(item.id)
		tenantID := tenant.ID(item.tenantID)
		if err := store.SaveInventory(ctx, inventory.Inventory{
			ID:       inventoryID,
			TenantID: inventory.TenantID(tenantID.String()),
			Name:     name,
		}); err != nil {
			t.Fatalf("save inventory: %v", err)
		}
		if item.owner != "" {
			if err := authorizer.GrantInventoryOwner(ctx, principal(item.owner), tenantID, inventoryID); err != nil {
				t.Fatalf("grant inventory owner: %v", err)
			}
		}
	}
}

func enqueuePoisonTenantOwnerOutboxEvent(t *testing.T, store *memory.Store, eventID string, tenantID string) {
	t.Helper()

	name, ok := tenant.NewName("Poison")
	if !ok {
		t.Fatalf("invalid poison tenant name")
	}
	auditID, ok := audit.NewID(eventID + "-audit")
	if !ok {
		t.Fatalf("invalid poison audit id")
	}
	record, ok := audit.NewRecord(
		auditID,
		audit.TenantID(tenantID),
		"",
		audit.PrincipalID("poison-user"),
		audit.ActionTenantCreated,
		audit.SourceAPI,
		audit.TargetTenant,
		tenantID,
		time.Now(),
		"",
		map[string]string{},
	)
	if !ok {
		t.Fatalf("invalid poison audit record")
	}
	if err := store.SaveTenantAndEnqueueOwnerGrant(context.Background(), eventID, tenant.Tenant{ID: tenant.ID(tenantID), Name: name}, identity.Principal{ID: identity.PrincipalID("poison-user")}, record); err != nil {
		t.Fatalf("enqueue poison outbox event: %v", err)
	}
}

func principal(id string) identity.Principal {
	return identity.Principal{ID: identity.PrincipalID(id)}
}

func performRequest(server *http.Server, method string, path string, authorization string, body any) *httptest.ResponseRecorder {
	return performRequestWithHeaders(server, method, path, authorization, nil, body)
}

func performRequestWithHeaders(server *http.Server, method string, path string, authorization string, headers map[string]string, body any) *httptest.ResponseRecorder {
	var requestBody []byte
	if body != nil {
		requestBody, _ = json.Marshal(body)
	}

	request := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	return response
}

func decodeBody(t *testing.T, response *httptest.ResponseRecorder, body any) {
	t.Helper()

	if err := json.NewDecoder(response.Body).Decode(body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

type errorResponse struct {
	Error struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Details []interface{} `json:"details"`
	} `json:"error"`
	Meta responseMeta `json:"meta"`
}

type responseMeta struct {
	RequestID  string          `json:"requestId,omitempty"`
	TenantID   string          `json:"tenantId,omitempty"`
	Pagination *paginationMeta `json:"pagination,omitempty"`
}

type paginationMeta struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
}

type seededState struct {
	tenants       []seedTenant
	inventories   []seedInventory
	ids           []string
	invitationTTL time.Duration
}

type seedTenant struct {
	id    string
	name  string
	owner string
}

type seedInventory struct {
	id       string
	tenantID string
	name     string
	owner    string
}

func assertSafeError(t *testing.T, response *httptest.ResponseRecorder, expectedCode string, expectedMessage string) {
	t.Helper()

	var body errorResponse
	decodeBody(t, response, &body)
	if body.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, body.Error.Code)
	}
	if body.Error.Message != expectedMessage {
		t.Fatalf("expected error message %q, got %q", expectedMessage, body.Error.Message)
	}
	if len(body.Error.Details) != 0 {
		t.Fatalf("expected no error details, got %+v", body.Error.Details)
	}
	if body.Meta.TenantID != "" || body.Meta.RequestID != "" {
		t.Fatalf("expected empty error metadata, got %+v", body.Meta)
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()

	var body errorResponse
	decodeBody(t, response, &body)
	if body.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, body.Error.Code)
	}
}

func paginationCursor(payload map[string]any) string {
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

type fakeIDGenerator struct {
	ids []string
}

func (f *fakeIDGenerator) NewID() string {
	if len(f.ids) == 0 {
		panic("fake ID generator exhausted")
	}
	id := f.ids[0]
	f.ids = f.ids[1:]
	return id
}

type fakeObserver struct {
	events []ports.Event
}

func (f *fakeObserver) Record(_ context.Context, event ports.Event) {
	f.events = append(f.events, event)
}
