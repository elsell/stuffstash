package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestHealthEndpointReturnsHealthyStatus(t *testing.T) {
	observer := &fakeObserver{}
	server := NewServer(":0", newTestApp(observer, "unused-id"))

	response := performRequest(server, http.MethodGet, "/healthz", "", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body struct {
		Service string `json:"service"`
		Status  string `json:"status"`
	}
	decodeBody(t, response, &body)

	if body.Service != string(app.ServiceNameStuffStash) {
		t.Fatalf("expected service %q, got %q", app.ServiceNameStuffStash, body.Service)
	}
	if body.Status != string(app.HealthStatusHealthy) {
		t.Fatalf("expected status %q, got %q", app.HealthStatusHealthy, body.Status)
	}

	if len(observer.events) != 1 {
		t.Fatalf("expected 1 observability event, got %d", len(observer.events))
	}
	if observer.events[0].Name != ports.EventHealthChecked {
		t.Fatalf("expected event %q, got %q", ports.EventHealthChecked, observer.events[0].Name)
	}
}

func TestProtectedEndpointsRejectMissingAndMalformedAuthentication(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "01ARZ3NDEKTSV4RRFFQ69G5FAV"))

	tests := []struct {
		name          string
		authorization string
	}{
		{name: "missing token"},
		{name: "malformed token", authorization: "Bearer nope"},
		{name: "empty principal", authorization: "Bearer dev:"},
		{name: "unsafe principal", authorization: "Bearer dev:user/one"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(server, http.MethodGet, "/me", test.authorization, nil)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d with body %s", http.StatusUnauthorized, response.Code, response.Body.String())
			}

			var body errorResponse
			decodeBody(t, response, &body)
			if body.Error.Code != "authentication_required" {
				t.Fatalf("expected authentication_required, got %q", body.Error.Code)
			}
		})
	}
}

func TestSecureTenantInventoryFlow(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"

	server := NewServer(":0", newTestApp(&fakeObserver{}, tenantID, inventoryID))

	me := performRequest(server, http.MethodGet, "/me", "Bearer dev:user-one", nil)
	if me.Code != http.StatusOK {
		t.Fatalf("expected /me status %d, got %d with body %s", http.StatusOK, me.Code, me.Body.String())
	}

	createTenant := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:user-one", map[string]string{"name": "Home"})
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected create tenant status %d, got %d with body %s", http.StatusCreated, createTenant.Code, createTenant.Body.String())
	}

	var tenantBody struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	decodeBody(t, createTenant, &tenantBody)
	if tenantBody.Data.ID != tenantID {
		t.Fatalf("expected tenant ID %q, got %q", tenantID, tenantBody.Data.ID)
	}

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:user-one", map[string]string{"name": "Tools"})
	if createInventory.Code != http.StatusCreated {
		t.Fatalf("expected create inventory status %d, got %d with body %s", http.StatusCreated, createInventory.Code, createInventory.Body.String())
	}

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:user-one", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}

	var listBody struct {
		Data []struct {
			ID       string `json:"id"`
			TenantID string `json:"tenantId"`
			Name     string `json:"name"`
		} `json:"data"`
	}
	decodeBody(t, list, &listBody)
	if len(listBody.Data) != 1 {
		t.Fatalf("expected 1 inventory, got %d", len(listBody.Data))
	}
	if listBody.Data[0].ID != inventoryID {
		t.Fatalf("expected inventory ID %q, got %q", inventoryID, listBody.Data[0].ID)
	}
}

func TestInventoryEndpointsDenyCrossUserAccess(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newTestApp(&fakeObserver{}, tenantID, "01ARZ3NDEKTSV4RRFFQ69G5FAW"))

	createTenant := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]string{"name": "Home"})
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected create tenant status %d, got %d with body %s", http.StatusCreated, createTenant.Code, createTenant.Body.String())
	}

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:other-user", map[string]string{"name": "Tools"})
	if createInventory.Code != http.StatusForbidden {
		t.Fatalf("expected create inventory status %d, got %d with body %s", http.StatusForbidden, createInventory.Code, createInventory.Body.String())
	}

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:other-user", nil)
	if list.Code != http.StatusForbidden {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusForbidden, list.Code, list.Body.String())
	}
}

func TestOpenAPIIsGenerated(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequest(server, http.MethodGet, "/openapi.json", "", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected OpenAPI status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"/tenants/{tenantId}/inventories"`)) {
		t.Fatalf("expected OpenAPI to include inventory path, got %s", response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"bearerAuth"`)) {
		t.Fatalf("expected OpenAPI to include bearer auth, got %s", response.Body.String())
	}
}

func newTestApp(observer ports.Observer, ids ...string) app.App {
	store := memory.NewStore()
	return app.New(app.Dependencies{
		Observer:    observer,
		Auth:        auth.NewLocalDevAuthenticator(),
		Authorizer:  memory.NewAuthorizer(),
		Tenants:     store,
		Inventories: store,
		IDs:         &fakeIDGenerator{ids: ids},
	})
}

func performRequest(server *http.Server, method string, path string, authorization string, body any) *httptest.ResponseRecorder {
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
		Code string `json:"code"`
	} `json:"error"`
}

type fakeIDGenerator struct {
	ids []string
}

func (f *fakeIDGenerator) NewID() string {
	if len(f.ids) == 0 {
		return "fixed-id"
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
