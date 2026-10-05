package httpapi

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAttachmentMetadataAndLifecycleBoundary(t *testing.T) {
	for _, action := range []string{"list", "show", "archive", "restore", "delete"} {
		t.Run(action, func(t *testing.T) {
			path := "/tenants/home/inventories/tools/assets/asset/attachments"
			method := "GET"
			if action != "list" {
				path += "/photo"
			}
			if action == "archive" || action == "restore" {
				path += "/" + action
				method = "PATCH"
			}
			if action == "delete" {
				method = "DELETE"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != method {
					t.Error("wrong HTTP method")
				}
				if action == "list" && (r.URL.Query().Get("limit") != "2" || r.URL.Query().Get("cursor") != "next") {
					t.Error("pagination lost")
				}
				if action == "delete" {
					w.WriteHeader(204)
					return
				}
				data := `{"id":"photo","tenantId":"home","inventoryId":"tools","assetId":"asset","fileName":"photo.jpg","contentType":"image/jpeg","sizeBytes":9007199254740993,"sha256":"digest","createdAt":"2026-10-05T00:00:00Z","lifecycleState":"active"}`
				if action == "list" {
					data = "[" + data + "]"
				}
				io.WriteString(w, `{"$schema":"schema","data":`+data+`,"meta":{"requestId":"trace"}}`)
			}))
			defer server.Close()
			for _, trial := range []struct {
				token, tenant, inventory, asset string
				allowed                         bool
			}{
				{"owner", "home", "tools", "asset", true}, {"", "home", "tools", "asset", false}, {"viewer", "home", "tools", "asset", false}, {"owner", "other", "tools", "asset", false}, {"owner", "home", "other", "asset", false}, {"owner", "home", "tools", "other", false},
			} {
				api, _ := New(server.URL, trial.token, server.Client())
				s := ports.Scope{Tenant: trial.tenant, Inventory: trial.inventory}
				var result any
				var err error
				switch action {
				case "list":
					result, err = api.Attachments(context.Background(), s, trial.asset, ports.Page{Limit: 2, Cursor: "next"})
				case "show":
					result, err = api.Attachment(context.Background(), s, trial.asset, "photo")
				case "delete":
					err = api.DeleteAttachment(context.Background(), s, trial.asset, "photo")
				default:
					result, err = api.ChangeAttachment(context.Background(), s, trial.asset, "photo", ports.AttachmentAction(action))
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
				if action == "delete" {
					continue
				}
				b, _ := json.Marshal(result)
				var envelope struct {
					Data json.RawMessage `json:"data"`
					Meta struct {
						RequestID string `json:"requestId"`
					} `json:"meta"`
				}
				if err := json.Unmarshal(b, &envelope); err != nil {
					t.Fatal(err)
				}
				var item ports.Attachment
				if action == "list" {
					var items []ports.Attachment
					json.Unmarshal(envelope.Data, &items)
					if len(items) != 1 {
						t.Fatal(string(b))
					}
					item = items[0]
				} else {
					json.Unmarshal(envelope.Data, &item)
				}
				if item.SizeBytes != 9007199254740993 || item.SHA256 != "digest" || item.FileName != "photo.jpg" || item.AssetID != "asset" || envelope.Meta.RequestID != "trace" {
					t.Fatalf("response lost fields: %s", b)
				}
			}
		})
	}
}
