package httpserver

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/app"
	"testing"
)

func TestExpirationReviewWireContractPreservesPrecision(t *testing.T) {
	value := realtimeActionPlanFromApp(app.RealtimeVoiceActionPlanProposal{PlanID: "plan", Commands: []app.RealtimeVoiceActionPlanCommand{{ID: "bottle", Kind: "create_asset", Expiration: &app.RealtimeVoiceActionPlanExpiration{Date: "2028-02", Precision: "month"}}}})
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		Commands []struct {
			Expiration *struct {
				Date      string
				Precision string
			}
		}
	}
	if err := json.Unmarshal(encoded, &output); err != nil {
		t.Fatal(err)
	}
	if len(output.Commands) != 1 || output.Commands[0].Expiration == nil || output.Commands[0].Expiration.Date != "2028-02" || output.Commands[0].Expiration.Precision != "month" {
		t.Fatal("wire contract lost expiration")
	}
}
