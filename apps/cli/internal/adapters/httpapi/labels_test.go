package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/labels"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

const labelInstance = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
const labelIdentity = "01ARZ3NDEKTSV4RRFFQ69G5FAW"

// labelServer keeps identities and rendered artifacts scoped like the API. Modes
// inject damaged content or a redirect after an otherwise successful render.
type labelServer struct {
	mu          sync.Mutex
	provisioned bool
	artifact    []byte
	mode        string
	resolutions int
}

func (f *labelServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Authorization") != "Bearer editor" {
		w.WriteHeader(403)
		return
	}
	response := func(data any) { _ = json.NewEncoder(w).Encode(map[string]any{"data": data, "meta": map[string]any{}}) }
	switch r.URL.Path {
	case "/instance":
		response(map[string]any{"protocolVersion": 1, "instanceId": labelInstance})
	case "/labels/v1/" + labelInstance + "/" + labelIdentity:
		f.resolutions++
		response(map[string]any{"instanceId": labelInstance, "labelId": labelIdentity, "tenantId": "home", "inventoryId": "garage", "assetId": "tool", "lifecycleState": "active"})
	case "/tenants/home/inventories/garage/assets/tool/label":
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		f.provisioned = true
		response(map[string]any{"labelId": labelIdentity})
	case "/tenants/home/inventories/garage/assets/tool/label-renders":
		if !f.provisioned || r.Method != "POST" {
			w.WriteHeader(409)
			return
		}
		var selection struct {
			Format string `json:"format"`
		}
		if json.NewDecoder(r.Body).Decode(&selection) != nil {
			w.WriteHeader(400)
			return
		}
		kind := "image/png"
		f.artifact = []byte("\x89PNG\r\n\x1a\nlabel")
		if selection.Format == "pdf" {
			kind = "application/pdf"
			f.artifact = []byte("%PDF-1.7\nlabel")
		}
		digest := sha256.Sum256(f.artifact)
		response(map[string]any{"id": "render-1", "contentType": kind, "sha256": hex.EncodeToString(digest[:]), "contentPath": "https://untrusted.invalid/steal-token"})
	case "/tenants/home/inventories/garage/label-renders/render-1/content":
		if f.artifact == nil {
			w.WriteHeader(404)
			return
		}
		if f.mode == "redirect" {
			http.Redirect(w, r, "https://untrusted.invalid/steal-token", 302)
			return
		}
		kind := "image/png"
		if strings.HasPrefix(string(f.artifact), "%PDF-") {
			kind = "application/pdf"
		}
		w.Header().Set("Content-Type", kind)
		if f.mode == "corrupt" {
			_, _ = w.Write([]byte("damaged"))
			return
		}
		_, _ = w.Write(f.artifact)
	default:
		w.WriteHeader(404)
	}
}
func TestLabelSDKKeepsScopeAndVerifiesArtifactBeforeReturningBytes(t *testing.T) {
	for _, format := range []string{"png", "pdf"} {
		t.Run(format, func(t *testing.T) {
			fake := &labelServer{}
			server := httptest.NewServer(fake)
			defer server.Close()
			client, _ := New(server.URL, "editor", server.Client())
			scope := ports.Scope{Tenant: "home", Inventory: "garage"}
			artifact, err := client.RenderLabel(context.Background(), scope, "tool", ports.LabelRenderSelection{Format: format})
			if err != nil || len(artifact.Content) == 0 || artifact.Format != format {
				t.Fatalf("render: %+v %v", artifact, err)
			}
			for _, mode := range []string{"corrupt", "redirect"} {
				fake.mu.Lock()
				fake.mode = mode
				fake.mu.Unlock()
				artifact, err = client.RenderLabel(context.Background(), scope, "tool", ports.LabelRenderSelection{Format: format})
				if err == nil || len(artifact.Content) != 0 {
					t.Fatalf("%s returned unverified bytes", mode)
				}
			}
			fake.mu.Lock()
			fake.mode = ""
			fake.mu.Unlock()
			for _, wrong := range []ports.Scope{{Tenant: "other", Inventory: "garage"}, {Tenant: "home", Inventory: "other"}} {
				if _, err = client.RenderLabel(context.Background(), wrong, "tool", ports.LabelRenderSelection{Format: format}); err == nil {
					t.Fatal("cross-scope render succeeded")
				}
			}
			for _, token := range []string{"", "expired", "viewer"} {
				denied, _ := New(server.URL, token, server.Client())
				if _, err = denied.RenderLabel(context.Background(), scope, "tool", ports.LabelRenderSelection{Format: format}); err == nil {
					t.Fatal("unauthorized render succeeded")
				}
			}
		})
	}
}
func TestLabelResolutionUsesConfiguredServerAndRejectsForeignInstances(t *testing.T) {
	fake := &labelServer{}
	server := httptest.NewServer(fake)
	defer server.Close()
	client, _ := New(server.URL, "editor", server.Client())
	ref, err := labels.Parse("https://old.invalid/prefix/l/v1/" + labelInstance + "/" + labelIdentity)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := client.ResolveLabel(context.Background(), ref)
	if err != nil || resolved.Data.InventoryID != "garage" {
		t.Fatalf("resolve: %+v %v", resolved, err)
	}
	ref.Instance = labelIdentity
	if _, err = client.ResolveLabel(context.Background(), ref); err == nil {
		t.Fatal("foreign instance accepted")
	}
	if fake.resolutions != 1 {
		t.Fatal("foreign instance reached resolver")
	}
	denied, _ := New(server.URL, "viewer", server.Client())
	ref.Instance = labelInstance
	if _, err = denied.ResolveLabel(context.Background(), ref); err == nil {
		t.Fatal("unauthorized resolution succeeded")
	}
}
