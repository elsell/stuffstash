package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImportSourceTransportPreservesBodyAndScope(t *testing.T) {
	const body = `{"sourceType":"legacy_homebox","baseUrl":"https://source.invalid","username":"source-user","password":"private-secret","includeImages":false,"allowPrivateNetwork":false,"allowInsecureTLS":false,"fileName":"backup.csv","contentBase64":"YQ=="}`
	for _, action := range []string{"preview", "start"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			broken := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				expected := "/tenants/home/inventories/garage/imports/jobs/preview"
				if action == "start" {
					expected = "/tenants/home/inventories/garage/imports/jobs/job/start"
				}
				if r.URL.Path != expected || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				raw, err := io.ReadAll(r.Body)
				if err != nil || string(raw) != body || r.Method != "POST" {
					t.Error("source request changed")
				}
				if broken {
					w.WriteHeader(503)
					io.WriteString(w, "private-secret")
					return
				}
				io.WriteString(w, `{"$schema":"job-schema","data":{"id":"job","status":"queued","counts":{"assets":9007199254740993},"progress":{"phase":"queued","done":0,"total":2}},"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			client, err := New(server.URL, "owner", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			call := func(scope ports.Scope) (ports.Result[ports.ImportJob], error) {
				if action == "preview" {
					return client.PreviewImportJob(context.Background(), scope, []byte(body))
				}
				return client.StartImportJob(context.Background(), scope, "job", []byte(body))
			}
			result, err := call(ports.Scope{Tenant: "home", Inventory: "garage"})
			if err != nil || result.Data.ID != "job" || result.Data.Status != "queued" || result.Data.Counts.Assets != 9007199254740993 || result.Meta == nil || result.Schema == nil {
				t.Fatalf("job mapping: %+v %v", result, err)
			}
			for _, scope := range []ports.Scope{{Tenant: "other", Inventory: "garage"}, {Tenant: "home", Inventory: "other"}} {
				if _, err := call(scope); err == nil {
					t.Fatal("scope bypass")
				}
			}
			broken = true
			before := calls
			if _, err := call(ports.Scope{Tenant: "home", Inventory: "garage"}); err == nil || strings.Contains(err.Error(), "private-secret") || calls != before+1 {
				t.Fatal("unsafe error or retry")
			}
		})
	}
}
