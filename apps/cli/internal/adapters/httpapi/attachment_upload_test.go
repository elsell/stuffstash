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

func TestAttachmentUploadTransport(t *testing.T) {
	for _, action := range []string{"create", "start", "complete"} {
		t.Run(action, func(t *testing.T) {
			body := `{"fileName":"file.pdf","contentType":"application/pdf","contentBase64":"AAH/"}`
			path := "/tenants/home/inventories/tools/assets/asset/attachments"
			if action == "start" {
				path += "/direct-uploads"
				body = `{"fileName":"file.pdf","contentType":"application/pdf","sizeBytes":9007199254740993}`
			}
			if action == "complete" {
				path += "/direct-uploads/upload/complete"
				body = ""
			}
			storageCalls := 0
			storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { storageCalls++ }))
			defer storage.Close()
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != "POST" {
					t.Error("wrong method")
				}
				b, _ := io.ReadAll(r.Body)
				if string(b) != body {
					t.Errorf("changed body: %s", b)
				}
				data := `{"id":"photo","tenantId":"home","inventoryId":"tools","assetId":"asset","fileName":"file.pdf","contentType":"application/pdf","sizeBytes":3,"sha256":"digest","createdAt":"today","lifecycleState":"active"}`
				if action == "start" {
					encoded, _ := json.Marshal(map[string]any{"uploadId": "upload", "attachmentId": "photo", "method": "POST", "url": storage.URL, "headers": map[string]string{"x-storage": "value"}, "formFields": map[string]string{"policy": "opaque"}, "expiresAt": "tomorrow"})
					data = string(encoded)
				}
				w.WriteHeader(201)
				io.WriteString(w, `{"$schema":"schema","data":`+data+`,"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			for _, trial := range []struct {
				token, tenant, inventory, asset string
				allowed                         bool
			}{{"owner", "home", "tools", "asset", true}, {"", "home", "tools", "asset", false}, {"viewer", "home", "tools", "asset", false}, {"owner", "other", "tools", "asset", false}, {"owner", "home", "other", "asset", false}, {"owner", "home", "tools", "other", false}} {
				api, _ := New(server.URL, trial.token, server.Client())
				scope := ports.Scope{Tenant: trial.tenant, Inventory: trial.inventory}
				before := calls
				var result any
				var err error
				switch action {
				case "create":
					result, err = api.CreateAttachment(context.Background(), scope, trial.asset, strings.NewReader(body))
				case "start":
					result, err = api.StartAttachmentUpload(context.Background(), scope, trial.asset, strings.NewReader(body))
				case "complete":
					result, err = api.CompleteAttachmentUpload(context.Background(), scope, trial.asset, "upload")
				}
				if calls != before+1 {
					t.Fatal("unexpected retry")
				}
				if !trial.allowed {
					if err == nil {
						t.Fatal("unauthorized request succeeded")
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				b, _ := json.Marshal(result)
				if !strings.Contains(string(b), `"requestId":"trace"`) || !strings.Contains(string(b), `"$schema":"schema"`) {
					t.Fatal(string(b))
				}
				if action == "start" {
					v := result.(ports.Result[ports.DirectUpload])
					if v.Data.Headers["x-storage"] != "value" || v.Data.FormFields["policy"] != "opaque" || v.Data.URL != storage.URL || v.Data.UploadID != "upload" || v.Data.AttachmentID != "photo" || v.Data.Method != "POST" || v.Data.ExpiresAt != "tomorrow" {
						t.Fatal(string(b))
					}
				}
			}
			if storageCalls != 0 {
				t.Fatal("API client visited upload URL")
			}
		})
	}
}
