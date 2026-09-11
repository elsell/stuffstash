package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestVoiceReviewIncludesOriginalExpiration(t *testing.T) {
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{}, &fakeIDGenerator{})
	command, err := application.realtimeVoiceActionPlanCommand(context.Background(), checkoutToolSession(), ports.ActionPlanCommandRecord{ID: "bottle", Kind: actionplan.CommandKindCreateAsset, ArgumentsJSON: []byte(`{"title":"Bottle","customAssetTypeId":"medicine","expiration":{"date":"2028-02","precision":"month"}}`)})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(command)
	var value struct {
		Expiration *struct {
			Date      string
			Precision string
		}
	}
	if err := json.Unmarshal(encoded, &value); err != nil {
		t.Fatal(err)
	}
	if value.Expiration == nil || value.Expiration.Date != "2028-02" || value.Expiration.Precision != "month" {
		t.Fatal("review dropped date or precision")
	}
}

func TestVoiceCorrectionReviewShowsSetAndClear(t *testing.T) {
	for _, clear := range []bool{false, true} {
		item := assetItem("bottle", "tenant-home", "inventory-home", asset.KindItem, "")
		application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{items: map[asset.ID]asset.Asset{item.ID: item}}, &fakeIDGenerator{})
		date := `{"date":"2028-02","precision":"month"}`
		if clear {
			date = "null"
		}
		command, err := application.realtimeVoiceActionPlanCommand(context.Background(), checkoutToolSession(), ports.ActionPlanCommandRecord{ID: "correction", Kind: actionplan.CommandKindUpdateAsset, ArgumentsJSON: []byte(`{"assetId":"bottle","expiration":` + date + `}`)})
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(command)
		var value struct {
			ExpirationCleared bool
			Expiration        *struct{ Date, Precision string }
		}
		_ = json.Unmarshal(encoded, &value)
		if command.Title != item.Title.String() || command.Operation != "update" || value.ExpirationCleared != clear {
			t.Fatal("correction review incomplete")
		}
		if clear {
			if value.Expiration != nil {
				t.Fatal("clear includes date")
			}
		} else if value.Expiration == nil || value.Expiration.Date != "2028-02" || value.Expiration.Precision != "month" {
			t.Fatal("correction dropped date")
		}
	}
}
