package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"net/http"
	"testing"
)

func TestNotificationPlacementUsesCurrentScopedAncestors(t *testing.T) {
	server, store := notificationHTTPFixture(t)
	ctx := context.Background()
	for _, id := range []string{"closet", "bin"} {
		title, _ := asset.NewTitle(id)
		parent := asset.ID("")
		if id == "bin" {
			parent = "closet"
		}
		if err := store.CreateAsset(ctx, asset.Asset{ID: asset.ID(id), TenantID: "home", InventoryID: "main", Title: title, Kind: asset.KindContainer, LifecycleState: asset.LifecycleStateActive, ParentAssetID: parent}, audit.Record{ID: audit.ID("seed-" + id)}, nil); err != nil {
			t.Fatal(err)
		}
	}
	item, _, _ := store.AssetByID(ctx, "home", "main", "bottle")
	item.ParentAssetID = "bin"
	if err := store.UpdateAsset(ctx, item, []audit.Record{{ID: "place"}}, nil); err != nil {
		t.Fatal(err)
	}
	const path = "/tenants/home/inventories/main/notifications/notice"
	response := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
	requireStatus(t, response, http.StatusOK)
	var body struct {
		Data struct {
			ParentTrail           []struct{ AssetID, Title string }
			ParentTrailIncomplete bool
		}
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data.ParentTrail) != 2 || body.Data.ParentTrail[0].Title != "closet" || body.Data.ParentTrail[1].AssetID != "bin" || body.Data.ParentTrailIncomplete {
		t.Fatalf("wrong placement: %+v", body.Data)
	}
	response = performRequest(server, http.MethodGet, path+"?parentAssetId=foreign&tenantId=other&inventoryId=other", "Bearer dev:owner", nil)
	requireStatus(t, response, http.StatusOK)
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data.ParentTrail) != 2 || body.Data.ParentTrail[1].AssetID != "bin" {
		t.Fatalf("query overrode placement scope: %+v", body.Data)
	}
	requireStatus(t, performRequest(server, http.MethodGet, "/tenants/other/inventories/main/notifications/notice", "Bearer dev:owner", nil), http.StatusNotFound)
	requireStatus(t, performRequest(server, http.MethodGet, path, "Bearer dev:viewer", nil), http.StatusNotFound)
}
