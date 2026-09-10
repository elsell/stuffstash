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

type liveInteractionInventory struct{ garage, office, kitchen, drill, officeScrewdriver, kitchenScrewdriver, coffeeCounter string }

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
	if strings.HasPrefix(scenario, "coffee-counter") {
		checkLiveCoffeeProposal(t, plan, fixture)
		return
	}
	if scenario == "missing-bedroom" {
		checkMissingBedroomProposal(t, plan)
		return
	}
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

func checkMissingBedroomProposal(t *testing.T, plan ports.ActionPlanRecord) {
	t.Helper()
	if len(plan.Commands) != 2 {
		t.Fatalf("expected destination and item, got %+v", plan.Commands)
	}
	var bedroomID string
	var item map[string]any
	for _, command := range plan.Commands {
		var args map[string]any
		if err := json.Unmarshal([]byte(command.ArgumentsJSON), &args); err != nil {
			t.Fatal(err)
		}
		title, _ := args["title"].(string)
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			t.Fatalf("unexpected command: %+v", command)
		}
		switch args["kind"] {
		case "location":
			if !strings.EqualFold(title, "Master Bedroom") {
				t.Fatalf("wrong destination: %+v", args)
			}
			bedroomID = command.ID
		case "item":
			if !strings.EqualFold(title, "Water Bottle") {
				t.Fatalf("wrong item: %+v", args)
			}
			item = args
		default:
			t.Fatalf("unexpected asset kind: %+v", args)
		}
		t.Logf("VOICE_INTERACTION_PROPOSAL kind=%s id=%s arguments=%s", command.Kind, command.ID, command.ArgumentsJSON)
	}
	if bedroomID == "" || item == nil || item["parentCommandId"] != bedroomID {
		t.Fatalf("missing dependent containment: bedroom=%s item=%+v", bedroomID, item)
	}
}

func seedLiveCoffeeCounter(t *testing.T, ctx context.Context, application app.App, kitchen string) string {
	t.Helper()
	result, err := application.CreateAssetWithOperation(ctx, app.CreateAssetInput{Principal: identity.Principal{ID: "user-1"}, Source: audit.SourceAPI, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: "container", Title: "Coffee Counter", ParentAssetID: kitchen})
	if err != nil {
		t.Fatal(err)
	}
	return result.Asset.ID.String()
}
func checkLiveCoffeeProposal(t *testing.T, plan ports.ActionPlanRecord, fixture liveInteractionInventory) {
	t.Helper()
	expected := 1
	if fixture.coffeeCounter == "" {
		expected = 2
	}
	if len(plan.Commands) != expected {
		t.Fatalf("expected %d commands, got %+v", expected, plan.Commands)
	}
	var item, counter map[string]any
	var counterCommand string
	for _, command := range plan.Commands {
		var args map[string]any
		if err := json.Unmarshal([]byte(command.ArgumentsJSON), &args); err != nil {
			t.Fatal(err)
		}
		title, _ := args["title"].(string)
		if command.Kind != actionplan.CommandKindCreateAsset {
			t.Fatalf("unexpected coffee operation: %+v", command)
		}
		if args["kind"] == "item" && strings.EqualFold(title, "Starbucks Coffee Beans") {
			item = args
		} else if args["kind"] == "container" && strings.EqualFold(title, "Coffee Counter") {
			counter = args
			counterCommand = command.ID
		} else {
			t.Fatalf("unexpected creation: %+v", args)
		}
		t.Logf("VOICE_INTERACTION_PROPOSAL kind=%s id=%s arguments=%s", command.Kind, command.ID, command.ArgumentsJSON)
	}
	if item == nil {
		t.Fatal("coffee beans missing")
	}
	if fixture.coffeeCounter != "" {
		if item["parentAssetId"] != fixture.coffeeCounter {
			t.Fatalf("wrong existing counter: %+v", item)
		}
	} else if counter == nil || counter["parentAssetId"] != fixture.kitchen || item["parentCommandId"] != counterCommand {
		t.Fatalf("lost kitchen/counter containment: counter=%+v item=%+v", counter, item)
	}
}
