package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestExpirationWorkspaceHTTPAccessAndFilters(t *testing.T) {
	server, _ := notificationHTTPFixture(t)
	base := "/tenants/home/inventories/main/expiration-assets"
	for _, token := range []string{"", "Bearer malformed", "Bearer dev:outsider"} {
		response := performRequest(server, http.MethodGet, base, token, nil)
		if response.Code != http.StatusUnauthorized && response.Code != http.StatusForbidden {
			t.Fatalf("denied caller returned %d: %s", response.Code, response.Body.String())
		}
	}
	for _, path := range []string{"/tenants/other/inventories/main/expiration-assets", "/tenants/home/inventories/other/expiration-assets"} {
		response := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
		if response.Code != http.StatusNotFound && response.Code != http.StatusForbidden {
			t.Fatalf("scope leak: %d", response.Code)
		}
	}
	for _, token := range []string{"Bearer dev:owner", "Bearer dev:viewer"} {
		response := performRequest(server, http.MethodGet, base+"?mode=all", token, nil)
		requireStatus(t, response, http.StatusOK)
		var result struct {
			Data struct {
				Items  []struct{ ID string }
				Counts struct{ All int }
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Data.Counts.All != 1 || len(result.Data.Items) != 1 || result.Data.Items[0].ID != "bottle" {
			t.Fatalf("wrong visible inventory: %s", response.Body.String())
		}
	}
	for _, query := range []string{"mode=bogus", "fromDate=2026-02-30", "fromDate=2026-09-30&throughDate=2026-09-01", "cursor=garbage"} {
		response := performRequest(server, http.MethodGet, base+"?"+query, "Bearer dev:owner", nil)
		if response.Code != http.StatusBadRequest && response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("invalid query returned %d", response.Code)
		}
	}
	response := performRequest(server, http.MethodGet, base+"?q=not-present", "Bearer dev:owner", nil)
	requireStatus(t, response, http.StatusOK)
	var empty struct {
		Data struct {
			Items  []any
			Counts struct{ All int }
		}
	}
	if err := json.Unmarshal(response.Body.Bytes(), &empty); err != nil {
		t.Fatal(err)
	}
	if len(empty.Data.Items) != 0 || empty.Data.Counts.All != 0 {
		t.Fatal("filtered counts include nonmatches")
	}
}

func TestExpirationWorkspaceHTTPCompletePagesAndCursorIsolation(t *testing.T) {
	server, store := notificationHTTPFixture(t)
	ctx := context.Background()
	original, _, err := store.AssetByID(ctx, "home", "main", "bottle")
	if err != nil {
		t.Fatal(err)
	}
	const total = 271
	for index := 0; index < total-1; index++ {
		item := original
		item.ID = asset.ID(fmt.Sprintf("package-%03d", index))
		item.Expiration, _ = expirationdate.ParseDate(fmt.Sprintf("2001-01-%02d", index%28+1), expirationdate.Day)
		if err := store.CreateAsset(ctx, item, audit.Record{ID: audit.ID(fmt.Sprintf("seed-package-%03d", index))}, nil); err != nil {
			t.Fatal(err)
		}
	}
	undated := original
	undated.ID = "undated"
	undated.Expiration = expirationdate.Date{}
	if err := store.CreateAsset(ctx, undated, audit.Record{ID: "seed-undated"}, nil); err != nil {
		t.Fatal(err)
	}
	const base = "/tenants/home/inventories/main/expiration-assets?mode=all&limit=17"
	cursor := ""
	seen := map[string]bool{}
	firstCursor := ""
	for page := 0; page < 30; page++ {
		response := performRequest(server, http.MethodGet, base+"&cursor="+url.QueryEscape(cursor), "Bearer dev:owner", nil)
		requireStatus(t, response, http.StatusOK)
		var body struct {
			Data struct {
				Items  []struct{ ID string }
				Counts struct{ All, Expired int }
			}
			Meta struct {
				Pagination struct {
					HasMore    bool
					NextCursor *string
				}
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Data.Counts.All != total || body.Data.Counts.Expired != total {
			t.Fatalf("partial counts: %s", response.Body.String())
		}
		for _, item := range body.Data.Items {
			if seen[item.ID] {
				t.Fatal("duplicate page item")
			}
			seen[item.ID] = true
		}
		if !body.Meta.Pagination.HasMore {
			break
		}
		if body.Meta.Pagination.NextCursor == nil || *body.Meta.Pagination.NextCursor == cursor {
			t.Fatal("no pagination progress")
		}
		cursor = *body.Meta.Pagination.NextCursor
		if firstCursor == "" {
			firstCursor = cursor
		}
	}
	if len(seen) != total {
		t.Fatalf("incomplete scan %d", len(seen))
	}
	for _, request := range []struct{ suffix, token string }{{"&q=bottle", "Bearer dev:owner"}, {"", "Bearer dev:viewer"}, {"&locationId=other", "Bearer dev:owner"}} {
		response := performRequest(server, http.MethodGet, base+"&cursor="+url.QueryEscape(firstCursor)+request.suffix, request.token, nil)
		if response.Code != http.StatusBadRequest && response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("cursor reused outside query/principal: %d", response.Code)
		}
	}
	kind, _, err := store.CustomAssetTypeByID(ctx, "home", "main", "medicine")
	if err != nil {
		t.Fatal(err)
	}
	kind.ExpirationEnabled = false
	if err := store.UpdateCustomAssetType(ctx, kind, audit.Record{ID: "disable-tracking"}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"all", "expired", "soon"} {
		response := performRequest(server, http.MethodGet, "/tenants/home/inventories/main/expiration-assets?mode="+mode, "Bearer dev:owner", nil)
		requireStatus(t, response, http.StatusOK)
		var body struct {
			Data struct {
				Items  []any
				Counts struct{ All, Soon, Expired int }
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Data.Counts.All != total || body.Data.Counts.Soon != 0 || body.Data.Counts.Expired != 0 {
			t.Fatal("tracking-off counts wrong")
		}
		if mode != "all" && len(body.Data.Items) != 0 {
			t.Fatal("tracking-off attention leaked")
		}
	}
}

func coverExpirationWorkspaceScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	server, _ := notificationHTTPFixture(t)
	token, status := "Bearer dev:owner", http.StatusOK
	if adversarial {
		token, status = "Bearer dev:outsider", http.StatusForbidden
	}
	coverage.request(t, server, http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/expiration-assets", "/tenants/home/inventories/main/expiration-assets", token, nil, status)
}

func TestExpirationWorkspaceHTTPBrowseAvailabilityAndArchive(t *testing.T) {
	server, _ := notificationHTTPFixture(t)
	base := "/tenants/home/inventories/main/expiration-assets?mode=all"
	count := func(query string, want int) {
		t.Helper()
		response := performRequest(server, http.MethodGet, base+query, "Bearer dev:viewer", nil)
		requireStatus(t, response, http.StatusOK)
		var result struct {
			Data struct {
				Items  []struct{ CurrentCheckout *struct{ ID string } }
				Counts struct{ All int }
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Data.Counts.All != want || len(result.Data.Items) != want {
			t.Fatalf("wrong filtered counts: %s", response.Body.String())
		}
		if query == "&kind=item&checkoutState=checked_out" && want == 1 && result.Data.Items[0].CurrentCheckout == nil {
			t.Fatal("missing checkout presentation")
		}
	}
	count("&kind=item&checkoutState=available", 1)
	count("&kind=container", 0)
	count("&checkoutState=checked_out", 0)
	response := performRequest(server, http.MethodPost, "/tenants/home/inventories/main/assets/bottle/checkout", "Bearer dev:owner", map[string]any{})
	requireStatus(t, response, http.StatusCreated)
	count("&kind=item&checkoutState=checked_out", 1)
	count("&checkoutState=available", 0)
	returned := performRequest(server, http.MethodPost, "/tenants/home/inventories/main/assets/bottle/return", "Bearer dev:owner", map[string]any{})
	requireStatus(t, returned, http.StatusOK)
	archived := performRequest(server, http.MethodPatch, "/tenants/home/inventories/main/assets/bottle/archive", "Bearer dev:owner", nil)
	requireStatus(t, archived, http.StatusOK)
	count("", 0)
	for _, query := range []string{"&kind=foreign", "&checkoutState=invalid"} {
		response := performRequest(server, http.MethodGet, base+query, "Bearer dev:viewer", nil)
		if response.Code != http.StatusUnprocessableEntity && response.Code != http.StatusBadRequest {
			t.Fatalf("invalid filter accepted: %s", response.Body.String())
		}
	}
}

func TestExpirationWorkspaceDenseCountsAndSettingsCursor(t *testing.T) {
	server, store := notificationHTTPFixture(t)
	ctx := context.Background()
	original, _, err := store.AssetByID(ctx, "home", "main", "bottle")
	if err != nil {
		t.Fatal(err)
	}
	const total = 4097
	for index := 1; index < total; index++ {
		item := original
		item.ID = asset.ID(fmt.Sprintf("dense-%05d", index))
		if err := store.CreateAsset(ctx, item, audit.Record{ID: audit.ID(fmt.Sprintf("dense-audit-%05d", index))}, nil); err != nil {
			t.Fatal(err)
		}
	}
	initialized := performRequest(server, http.MethodPost, "/tenants/home/inventories/main/notification-preferences/initialize", "Bearer dev:owner", map[string]any{"timezone": "UTC"})
	requireStatus(t, initialized, http.StatusOK)
	started := time.Now()
	response := performRequest(server, http.MethodGet, "/tenants/home/inventories/main/expiration-assets?mode=expired&limit=30", "Bearer dev:owner", nil)
	requireStatus(t, response, http.StatusOK)
	var result struct {
		Data struct {
			Items  []struct{ ID string }
			Counts struct{ All, Expired int }
		}
		Meta struct {
			Pagination struct {
				NextCursor string
				HasMore    bool
			}
		}
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Data.Counts.All != total || result.Data.Counts.Expired != total || len(result.Data.Items) != 30 || !result.Meta.Pagination.HasMore {
		t.Fatalf("incomplete dense result: %s", response.Body.String())
	}
	t.Logf("Complete counts across %d dated assets and 30 result rows: %s", total, time.Since(started))
	changed := performRequest(server, http.MethodPut, "/tenants/home/inventories/main/notification-preferences", "Bearer dev:owner", map[string]any{"revision": 1, "defaults": map[string]any{"enabled": false, "upcoming": false, "expired": false, "advanceDays": 7}, "timezone": "America/Los_Angeles", "pushEnabled": false})
	requireStatus(t, changed, http.StatusOK)
	stale := performRequest(server, http.MethodGet, "/tenants/home/inventories/main/expiration-assets?mode=expired&limit=30&cursor="+url.QueryEscape(result.Meta.Pagination.NextCursor), "Bearer dev:owner", nil)
	requireStatus(t, stale, http.StatusBadRequest)
	fresh := performRequest(server, http.MethodGet, "/tenants/home/inventories/main/expiration-assets?mode=expired&limit=30", "Bearer dev:owner", nil)
	requireStatus(t, fresh, http.StatusOK)
	if err := json.Unmarshal(fresh.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Data.Counts.Expired != total {
		t.Fatal("notification delivery switches hid expired inventory facts")
	}
}
