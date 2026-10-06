package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArchiveJobTransportScopeAndFields(t *testing.T) {
	for _, action := range []string{"list", "show", "preview", "retry", "delete"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			expectedInventory := "garage"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				path := "/tenants/home/archive-jobs"
				method := "GET"
				if action != "list" {
					path += "/job"
				}
				if action == "preview" || action == "retry" {
					path += "/" + action
				}
				if action == "retry" {
					method = "POST"
				}
				if action == "delete" {
					method = "DELETE"
				}
				if r.URL.Path != path || r.URL.Query().Get("inventoryId") != expectedInventory || r.Header.Get("Authorization") != "Bearer owner" {
					w.WriteHeader(403)
					return
				}
				if r.Method != method {
					t.Error("wrong method")
				}
				if action == "delete" {
					w.WriteHeader(204)
					return
				}
				data := `{"id":"job","createdAt":"2026-10-05T12:00:00Z","expiresAt":"2026-10-06T12:00:00Z","inventoryId":"garage","destinationInventoryId":"new","failure":"retryable","kind":"export","otherFiles":false,"photos":true,"phase":"pack","state":"queued"}`
				if action == "list" {
					if r.URL.Query().Get("after") != "previous" || r.URL.Query().Get("limit") != "2" {
						t.Error("lost paging")
					}
					data = "[" + data + "]"
				}
				if action == "preview" {
					data = `{"assets":9007199254740993,"customAssetTypes":2,"customFields":3,"inventoryName":"Home","keyRemappings":[{"sourceKey":"old","destinationKey":"new","family":"tag"}],"omittedAttachments":4,"otherFiles":5,"photos":6,"tags":7}`
				}
				io.WriteString(w, `{"$schema":"schema","data":`+data+`,"meta":{"requestId":"trace","pagination":{"limit":2,"hasMore":true,"nextCursor":"next"}}}`)
			}))
			defer server.Close()
			client, err := New(server.URL, "owner", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			call := func(s ports.Scope) error {
				switch action {
				case "list":
					r, e := client.ArchiveJobs(context.Background(), s, ports.Page{Limit: 2, Cursor: "previous"})
					if e == nil && (len(r.Data) != 1 || r.Data[0].DestinationInventoryID == nil || r.Pagination == nil) {
						t.Error("lost list fields")
					}
					return e
				case "preview":
					r, e := client.ArchivePreview(context.Background(), s, "job")
					if e == nil && (r.Data.Assets != 9007199254740993 || len(r.Data.KeyRemappings) != 1 || r.Data.KeyRemappings[0].DestinationKey != "new") {
						t.Error("lost preview")
					}
					return e
				case "delete":
					return client.DeleteArchiveJob(context.Background(), s, "job")
				default:
					var r ports.Result[ports.ArchiveJob]
					var e error
					if action == "retry" {
						r, e = client.RetryArchiveJob(context.Background(), s, "job")
					} else {
						r, e = client.ArchiveJob(context.Background(), s, "job")
					}
					if e == nil && (r.Data.ID != "job" || !r.Data.Photos || r.Data.OtherFiles || r.Data.Failure == nil || r.Meta == nil || r.Schema == nil) {
						t.Error("lost job")
					}
					return e
				}
			}
			if err := call(ports.Scope{Tenant: "home", Inventory: "garage"}); err != nil {
				t.Fatal(err)
			}
			for _, s := range []ports.Scope{{Tenant: "other", Inventory: "garage"}, {Tenant: "home", Inventory: "other"}} {
				if err := call(s); err == nil {
					t.Fatal("scope bypass")
				}
			}
			expectedInventory = ""
			if err := call(ports.Scope{Tenant: "home"}); err != nil {
				t.Fatal("household-only archive request failed", err)
			}
			if calls != 4 {
				t.Fatal("unexpected retry")
			}
		})
	}
}
