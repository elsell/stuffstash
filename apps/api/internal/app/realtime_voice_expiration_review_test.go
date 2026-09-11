package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
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
