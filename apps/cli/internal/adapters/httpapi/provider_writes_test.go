package httpapi

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProviderWritesKeepContractAndProtectCredentials(t *testing.T) {
	for _, action := range []string{"create", "update", "credential"} {
		t.Run(action, func(t *testing.T) {
			body := `{"capability":"language_inference","providerKind":"local_http","displayName":"Local","enable":false,"endpointUrl":"https://model.example","modelName":"m","promptTemplate":"","runtimeOptions":{"limit":9007199254740993},"capabilityMetadata":{"tools":false}}`
			method, path := "POST", "/tenants/home/provider-profiles"
			if action == "update" {
				method = "PATCH"
				path += "/profile"
				body = `{"displayName":"Updated","endpointUrl":"","modelName":"","promptTemplate":"","runtimeOptions":{"limit":9007199254740993},"capabilityMetadata":{"tools":false}}`
			}
			if action == "credential" {
				method = "PUT"
				path += "/profile/credential"
				body = `{"purpose":"api_key","credential":"private-provider-secret"}`
			}
			status, calls := 200, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Header.Get("Authorization") != "Bearer owner" || r.URL.Path != path {
					w.WriteHeader(403)
					return
				}
				if r.Method != method || r.URL.RawQuery != "" || r.Header.Get("Content-Type") != "application/json" {
					t.Error("wrong request")
				}
				raw, err := io.ReadAll(r.Body)
				if err != nil || string(raw) != body {
					t.Error("changed JSON input")
				}
				if status != 200 {
					w.WriteHeader(status)
					io.WriteString(w, "private-provider-secret")
					return
				}
				io.WriteString(w, `{"data":{"id":"profile","tenantId":"home","credentialStatus":"configured","runtimeOptions":{"limit":9007199254740993},"credential":"private-provider-secret"},"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			client, err := New(server.URL, "owner", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			call := func(tenant string) (ports.Result[ports.ProviderProfile], error) {
				switch action {
				case "create":
					return client.CreateProvider(context.Background(), tenant, []byte(body))
				case "update":
					return client.UpdateProvider(context.Background(), tenant, "profile", []byte(body))
				default:
					return client.ReplaceProviderCredential(context.Background(), tenant, "profile", []byte(body))
				}
			}
			result, err := call("home")
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(result)
			if err != nil || !strings.Contains(string(raw), `"limit":9007199254740993`) || result.Data.CredentialStatus != "configured" || result.Meta == nil || strings.Contains(string(raw), "private-provider-secret") {
				t.Fatal("lost response contract or leaked credential")
			}
			if _, err := call("other"); err == nil {
				t.Fatal("household bypass")
			}
			for _, code := range []int{401, 403, 409, 503} {
				status = code
				before := calls
				if _, err := call("home"); err == nil || calls != before+1 || strings.Contains(err.Error(), "private-provider-secret") {
					t.Fatal("unsafe error or retry")
				}
			}
		})
	}
}
