package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
)

func TestCompatibleProviderDiagnosticAuthorizesBeforeNetwork(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer raw:") {
			t.Error("credential vault result missing")
		}
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"tool_calls","message":{"tool_calls":[{"id":"probe","type":"function","function":{"name":"ready","arguments":"{\"status\":\"ready\"}"}}]}}]}`))
	}))
	defer upstream.Close()
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	tester := voice.NewProviderProfileTester(voice.ProviderProfileFactory{CompatibleEndpoints: []string{upstream.URL}})
	server := NewServer(":0", newProviderProfileTestAppWithTester(t, seededState{tenants: []seedTenant{{id: tenantID, name: "Home", owner: "owner"}, {id: otherID, name: "Other", owner: "other"}}, ids: []string{"01ARZ3NDEKTSV4RRFFQ69G5FAW", "audit-create", "audit-config", "credential", "audit-credential", "audit-test"}}, tester))
	create := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/provider-profiles", "Bearer dev:owner", map[string]any{"capability": "language_inference", "providerKind": "local_http", "displayName": "Local", "endpointUrl": upstream.URL, "modelName": "local-test", "runtimeOptions": map[string]any{}, "capabilityMetadata": map[string]any{}})
	if create.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", create.Code, create.Body.String())
	}
	id := decodeProviderProfile(t, create).Data.ID
	path := "/tenants/" + tenantID + "/provider-profiles/" + id
	credential := performRequest(server, http.MethodPut, path+"/credential", "Bearer dev:owner", map[string]string{"purpose": "api_key", "credential": "test-secret"})
	if credential.Code != http.StatusOK {
		t.Fatalf("credential: %d %s", credential.Code, credential.Body.String())
	}
	for _, attempt := range []struct {
		token, path string
		status      int
	}{
		{"", path, http.StatusUnauthorized},
		{"Bearer dev:stranger", path, http.StatusForbidden},
		{"Bearer dev:other", path, http.StatusForbidden},
		{"Bearer dev:other", "/tenants/" + otherID + "/provider-profiles/" + id, http.StatusNotFound},
	} {
		result := performRequest(server, http.MethodPost, attempt.path+"/test", attempt.token, map[string]any{})
		if result.Code != attempt.status || calls.Load() != 0 {
			t.Fatalf("unauthorized network diagnostic: got %d want %d calls=%d", result.Code, attempt.status, calls.Load())
		}
	}

	result := performRequest(server, http.MethodPost, path+"/test", "Bearer dev:owner", map[string]any{})
	if result.Code != http.StatusOK || calls.Load() != 1 || strings.Contains(result.Body.String(), "raw:") || strings.Contains(result.Body.String(), "test-secret") {
		t.Fatalf("owner diagnostic: %d calls=%d %s", result.Code, calls.Load(), result.Body.String())
	}
}
