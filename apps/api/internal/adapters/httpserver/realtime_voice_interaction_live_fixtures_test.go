package httpserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type liveInteractionInventory struct{ garage, office, kitchen, drill, officeScrewdriver, kitchenScrewdriver string }

func seedLiveInteractionInventory(t *testing.T, ctx context.Context, application app.App) liveInteractionInventory {
	t.Helper()
	create := func(kind, title, parent, description string) string {
		t.Helper()
		result, err := application.CreateAssetWithOperation(ctx, app.CreateAssetInput{Principal: identity.Principal{ID: "user-1"}, Source: audit.SourceAPI, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: kind, Title: title, ParentAssetID: parent, Description: description})
		if err != nil {
			t.Fatal(err)
		}
		return result.Asset.ID.String()
	}
	fixture := liveInteractionInventory{garage: create("location", "Garage", "", ""), office: create("location", "Office", "", ""), kitchen: create("location", "Kitchen", "", "")}
	fixture.drill = create("item", "Cordless drill", fixture.office, "Red cordless drill.")
	fixture.officeScrewdriver = create("item", "Screwdriver", fixture.office, "Yellow handle.")
	fixture.kitchenScrewdriver = create("item", "Screwdriver", fixture.kitchen, "Green handle.")
	return fixture
}
func checkLiveInteractionProposal(t *testing.T, scenario string, plan ports.ActionPlanRecord, fixture liveInteractionInventory) {
	t.Helper()
	if plan.State != actionplan.StateProposed {
		t.Fatalf("unapproved plan is not proposed: %s", plan.State)
	}
	expectedCount := 1
	if scenario == "dependent-move" {
		expectedCount = 2
	}
	if len(plan.Commands) != expectedCount {
		t.Fatalf("expected %d useful commands, got %+v", expectedCount, plan.Commands)
	}
	type command struct {
		id        string
		arguments map[string]any
	}
	byKind := map[string]command{}
	for _, c := range plan.Commands {
		kind := string(c.Kind)
		if _, duplicate := byKind[kind]; duplicate {
			t.Fatalf("duplicated operation %s", kind)
		}
		var arguments map[string]any
		if err := json.Unmarshal([]byte(c.ArgumentsJSON), &arguments); err != nil {
			t.Fatal(err)
		}
		byKind[kind] = command{id: c.ID, arguments: arguments}
		t.Logf("VOICE_INTERACTION_PROPOSAL kind=%s id=%s arguments=%s", kind, c.ID, c.ArgumentsJSON)
	}
	switch scenario {
	case "move-existing", "ambiguous-move":
		target := fixture.drill
		if scenario == "ambiguous-move" {
			target = fixture.kitchenScrewdriver
		}
		move, found := byKind["move_asset"]
		if !found || move.arguments["assetId"] != target || move.arguments["parentAssetId"] != fixture.garage {
			t.Fatalf("existing target or destination not resolved: %+v", byKind)
		}
	case "create-additional":
		create, found := byKind["create_asset"]
		title, _ := create.arguments["title"].(string)
		if !found || !strings.Contains(strings.ToLower(title), "cordless drill") || create.arguments["kind"] != "item" || create.arguments["parentAssetId"] != fixture.garage {
			t.Fatalf("explicitly additional item not proposed correctly: %+v", byKind)
		}
	case "dependent-move":
		create, created := byKind["create_asset"]
		move, moved := byKind["move_asset"]
		title, _ := create.arguments["title"].(string)
		if !created || !moved || !strings.Contains(strings.ToLower(title), "blue") || !strings.Contains(strings.ToLower(title), "toolbox") || create.arguments["kind"] != "container" || create.arguments["parentAssetId"] != fixture.garage || move.arguments["assetId"] != fixture.drill || move.arguments["parentCommandId"] != create.id {
			t.Fatalf("dependent destination or existing drill lost: %+v", byKind)
		}
	default:
		t.Fatalf("unknown proposal scenario %s", scenario)
	}
}
