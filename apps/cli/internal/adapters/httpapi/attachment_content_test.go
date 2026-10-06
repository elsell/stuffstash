package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAttachmentContentBoundary(t *testing.T) {
	leaked := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked = true }))
	defer target.Close()
	for _, thumbnail := range []bool{false, true} {
		for _, status := range []int{200, 401, 403, 404, 302, 206} {
			t.Run(http.StatusText(status)+map[bool]string{true: " thumbnail", false: " content"}[thumbnail], func(t *testing.T) {
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					suffix := "content"
					if thumbnail {
						suffix = "thumbnail"
					}
					if r.URL.Path != "/tenants/home/inventories/tools/assets/asset/attachments/photo/"+suffix || r.Header.Get("Authorization") != "Bearer owner" {
						w.WriteHeader(403)
						return
					}
					if thumbnail && r.URL.Query().Get("variant") != "large" {
						t.Error("variant lost")
					}
					w.Header().Set("Content-Type", "image/png")
					w.Header().Set("Content-Disposition", `attachment; filename="../../outside.png"`)
					w.Header().Set("Location", target.URL)
					w.WriteHeader(status)
					w.Write([]byte{0, 1, 255, 2})
				}))
				defer server.Close()
				api, _ := New(server.URL, "owner", server.Client())
				for _, scope := range []ports.Scope{{Tenant: "home", Inventory: "tools"}, {Tenant: "other", Inventory: "tools"}, {Tenant: "home", Inventory: "other"}} {
					var result ports.AttachmentContent
					var err error
					if thumbnail {
						result, err = api.AttachmentThumbnail(context.Background(), scope, "asset", "photo", "large")
					} else {
						result, err = api.AttachmentContent(context.Background(), scope, "asset", "photo")
					}
					success := status == 200 && scope.Tenant == "home" && scope.Inventory == "tools"
					if !success {
						if err == nil {
							result.Body.Close()
							t.Fatal("unsuccessful response exposed as file")
						}
						continue
					}
					if err != nil {
						t.Fatal(err)
					}
					data, err := io.ReadAll(result.Body)
					result.Body.Close()
					if err != nil || string(data) != string([]byte{0, 1, 255, 2}) || result.ContentType != "image/png" || result.ContentDisposition != `attachment; filename="../../outside.png"` {
						t.Fatal("binary response changed")
					}
				}
				before := calls
				if _, err := api.AttachmentThumbnail(context.Background(), ports.Scope{Tenant: "home", Inventory: "tools"}, "asset", "photo", "huge"); err == nil || calls != before {
					t.Fatal("invalid variant sent")
				}
			})
		}
	}
	if leaked {
		t.Fatal("redirect followed")
	}
}
