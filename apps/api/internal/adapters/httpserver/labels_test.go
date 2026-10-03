package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestLabelProvisionResolveRenderSecurity(t *testing.T) {
	server, _, az := labelTestServer(t)
	viewer := identity.Principal{ID: "viewer"}
	if err := az.GrantInventoryViewer(context.Background(), viewer, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"item", "container", "location"} {
		created := performRequest(server, http.MethodPost, labelPrefix+"/assets", "Bearer dev:owner", map[string]any{"title": "Café tools", "kind": kind})
		if created.Code != 201 {
			t.Fatalf("create: %d %s", created.Code, created.Body.String())
		}
		assetID := decodeAsset(t, created).Data.ID
		path := labelPrefix + "/assets/" + assetID + "/label"
		absent := performRequest(server, "GET", path, "Bearer dev:viewer", nil)
		if absent.Code != 404 {
			t.Fatalf("GET must not provision: %d", absent.Code)
		}
		for _, token := range []string{"", "Bearer invalid", "Bearer dev:other"} {
			res := performRequest(server, "POST", path, token, nil)
			if res.Code != 401 && res.Code != 403 {
				t.Fatalf("provision leaked: %d %s", res.Code, res.Body.String())
			}
		}
		provisioned := performRequest(server, "POST", path, "Bearer dev:viewer", nil)
		if provisioned.Code != 200 {
			t.Fatalf("provision: %d %s", provisioned.Code, provisioned.Body.String())
		}
		label := labelResponseData(t, provisioned.Body.Bytes())
		repeat := performRequest(server, "POST", path, "Bearer dev:viewer", nil)
		again := labelResponseData(t, repeat.Body.Bytes())
		if again["labelId"] != label["labelId"] {
			t.Fatal("provision not idempotent")
		}
		link := label["url"].(string)
		if !strings.HasPrefix(link, "https://example.test/stash/l/v1/") {
			t.Fatal(link)
		}
		resolver := "/labels/v1/" + label["instanceId"].(string) + "/" + label["labelId"].(string)
		resolved := performRequest(server, "GET", resolver, "Bearer dev:viewer", nil)
		if resolved.Code != 200 {
			t.Fatal(resolved.Body.String())
		}
		denied := performRequest(server, "GET", resolver, "Bearer dev:other", nil)
		missing := performRequest(server, "GET", "/labels/v1/01ARZ3NDEKTSV4RRFFQ69G5FAZ/01ARZ3NDEKTSV4RRFFQ69G5FAZ", "Bearer dev:other", nil)
		if denied.Code != 404 || missing.Code != 404 || denied.Body.String() != missing.Body.String() {
			t.Fatal("resolver leaks identities")
		}
		media := printing.MediaSnapshot{WidthMicrometers: 29000, HeightMicrometers: 90000, MarginsMicrometers: printing.Margins{Left: 1546, Right: 1546, Top: 3047, Bottom: 3048}, ResolutionDPI: 300, RasterWidth: 306, RasterHeight: 991, Orientation: printing.OrientationFeed, ColorMode: printing.ColorMonochrome, CutPolicy: printing.CutNone, DisplayRotation: 270}
		body := map[string]any{"media": media, "template": printing.TemplateSelection{ID: printing.TemplateQRTitle, Version: 1, Options: printing.TemplateOptions{ShowReference: true}}, "format": "png"}
		render := performRequest(server, "POST", path+"-renders", "Bearer dev:viewer", body)
		if render.Code != 201 {
			t.Fatalf("render: %d %s", render.Code, render.Body.String())
		}
		artifact := labelResponseData(t, render.Body.Bytes())
		contentPath := artifact["contentPath"].(string)
		content := performRequest(server, "GET", contentPath, "Bearer dev:viewer", nil)
		if content.Code != 200 || content.Header().Get("Cache-Control") != "private, no-store" || content.Header().Get("Content-Type") != "image/png" {
			t.Fatalf("content: %d %s", content.Code, content.Body.String())
		}
		otherPath := strings.Replace(contentPath, labelInventory, labelOtherInventory, 1)
		cross := performRequest(server, "GET", otherPath, "Bearer dev:other", nil)
		if cross.Code != 404 {
			t.Fatalf("cross-scope content: %d", cross.Code)
		}
		if err := az.RevokeInventoryViewer(context.Background(), viewer, labelTenant, inventory.InventoryID(labelInventory)); err != nil {
			t.Fatal(err)
		}
		revoked := performRequest(server, "GET", contentPath, "Bearer dev:viewer", nil)
		if revoked.Code != 403 {
			t.Fatalf("content after revoke: %d", revoked.Code)
		}
		if err := az.GrantInventoryViewer(context.Background(), viewer, labelTenant, labelInventory); err != nil {
			t.Fatal(err)
		}

		for _, token := range []string{"", "Bearer invalid"} {
			denied := performRequest(server, "GET", contentPath, token, nil)
			if denied.Code != 401 {
				t.Fatalf("unauthenticated content: %d", denied.Code)
			}
		}
		archived := performRequest(server, "PATCH", labelPrefix+"/assets/"+assetID+"/archive", "Bearer dev:owner", nil)
		if archived.Code != 200 {
			t.Fatalf("archive: %d %s", archived.Code, archived.Body.String())
		}
		archivedScan := performRequest(server, "GET", resolver, "Bearer dev:viewer", nil)
		if archivedScan.Code != 200 || labelResponseData(t, archivedScan.Body.Bytes())["lifecycleState"] != "archived" {
			t.Fatal("archived label no longer resolves")
		}
		deleted := performRequest(server, "DELETE", labelPrefix+"/assets/"+assetID, "Bearer dev:owner", nil)
		if deleted.Code != 204 {
			t.Fatalf("delete: %d %s", deleted.Code, deleted.Body.String())
		}
		deletedScan := performRequest(server, "GET", resolver, "Bearer dev:viewer", nil)
		if deletedScan.Code != 404 {
			t.Fatal("deleted label resolves")
		}
		deletedContent := performRequest(server, "GET", contentPath, "Bearer dev:viewer", nil)
		if deletedContent.Code != 404 {
			t.Fatal("deleted label artifact remains exposed")
		}
	}
}

func TestLabelContentExpiresBeforeCleanup(t *testing.T) {
	clock := &labelTestClock{now: time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)}
	server, _, _ := labelTestServer(t, clock)
	created := performRequest(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Tools"})
	if created.Code != 201 {
		t.Fatal(created.Body.String())
	}
	assetID := decodeAsset(t, created).Data.ID
	path := labelPrefix + "/assets/" + assetID + "/label"
	provisioned := performRequest(server, "POST", path, "Bearer dev:owner", nil)
	if provisioned.Code != 200 {
		t.Fatal(provisioned.Body.String())
	}
	media := printing.MediaSnapshot{WidthMicrometers: 29000, HeightMicrometers: 90000, MarginsMicrometers: printing.Margins{Left: 1546, Right: 1546, Top: 3047, Bottom: 3048}, ResolutionDPI: 300, RasterWidth: 306, RasterHeight: 991, Orientation: printing.OrientationFeed, ColorMode: printing.ColorMonochrome, CutPolicy: printing.CutNone, DisplayRotation: 270}
	render := performRequest(server, "POST", path+"-renders", "Bearer dev:owner", map[string]any{"media": media, "template": printing.TemplateSelection{ID: printing.TemplateQROnly, Version: 1}, "format": "pdf"})
	if render.Code != 201 {
		t.Fatal(render.Body.String())
	}
	contentPath := labelResponseData(t, render.Body.Bytes())["contentPath"].(string)
	before := performRequest(server, "GET", contentPath, "Bearer dev:owner", nil)
	if before.Code != 200 || before.Header().Get("Content-Type") != "application/pdf" {
		t.Fatal("PDF download failed")
	}
	clock.now = clock.now.Add(time.Hour)
	expired := performRequest(server, "GET", contentPath, "Bearer dev:owner", nil)
	if expired.Code != 404 {
		t.Fatal("expired artifact exposed before cleanup")
	}
}
