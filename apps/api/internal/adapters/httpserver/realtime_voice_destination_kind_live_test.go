package httpserver

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Real model decisions through the authenticated realtime endpoint. Text input
// isolates destination classification from transcription; this is not device evidence.
func TestGoogleLiveDestinationKindCorpus(t *testing.T) {
	if os.Getenv("STUFF_STASH_GOOGLE_LIVE_TESTS") != "1" {
		t.Skip("explicit live Google opt-in required")
	}
	for _, scenario := range []struct{ name, text string }{
		{"missing-bin", "Place my cordless drill in Bin 60."},
		{"missing-room", "Place my cordless drill in the master bedroom."},
		{"existing-bin-location", "Place my cordless drill in Bin 60."},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			runLiveInteraction(t, scenario.name, nil, scenario.text)
		})
	}
}

func checkLiveDestinationKindProposal(t *testing.T, scenario string, plan ports.ActionPlanRecord, fixture liveInteractionInventory) {
	t.Helper()
	expected := 2
	if scenario == "existing-bin-location" {
		expected = 1
	}
	if len(plan.Commands) != expected {
		t.Fatalf("expected %d commands, got %+v", expected, plan.Commands)
	}
	var destinationCommand string
	for i, command := range plan.Commands {
		var args map[string]any
		if err := json.Unmarshal(command.ArgumentsJSON, &args); err != nil {
			t.Fatal(err)
		}
		t.Logf("DESTINATION_PROPOSAL kind=%s arguments=%s", command.Kind, command.ArgumentsJSON)
		if expected == 2 && i == 0 {
			kind, title := "container", "Bin 60"
			if scenario == "missing-room" {
				kind, title = "location", "Master Bedroom"
			}
			actualKind := args["kind"]
			if command.Kind == actionplan.CommandKindCreateLocation {
				actualKind = "location"
			} else if command.Kind != actionplan.CommandKindCreateAsset {
				t.Fatalf("expected destination creation: %+v", command)
			}
			actualTitle, _ := args["title"].(string)
			if actualKind != kind || !strings.EqualFold(actualTitle, title) {
				t.Fatalf("wrong destination kind or title: %+v", args)
			}
			destinationCommand = command.ID
			continue
		}
		if command.Kind != actionplan.CommandKindMoveAsset || args["assetId"] != fixture.drill {
			t.Fatalf("expected move of existing drill: %+v", command)
		}
		if expected == 1 && args["parentAssetId"] != fixture.bin || expected == 2 && args["parentCommandId"] != destinationCommand {
			t.Fatalf("wrong destination reference: %+v", args)
		}
	}
}
