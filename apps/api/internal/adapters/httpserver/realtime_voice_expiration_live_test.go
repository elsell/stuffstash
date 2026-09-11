package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"nhooyr.io/websocket"
)

func TestGoogleLiveExpirationConversationCorpus(t *testing.T) {
	if os.Getenv("STUFF_STASH_GOOGLE_LIVE_TESTS") != "1" {
		t.Skip("explicit live Google opt-in required")
	}
	_ = liveVoiceRequired(t, "STUFF_STASH_VOICE_EVIDENCE_DIR")
	now := time.Now().UTC()
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
	for _, scenario := range []struct {
		name, text, date, precision string
		write                       bool
	}{
		{"add-day", "Add a bottle of aspirin to Bin 8, using the Medicine type, expiring February 29, 2028.", "2028-02-29", "day", true},
		{"add-month", "Add a bottle of aspirin to Bin 8, using the Medicine type, expiring February 2028.", "2028-02", "month", true},
		{"add-relative", "Add a bottle of aspirin to Bin 8, using the Medicine type, expiring next month.", nextMonth, "month", true},
		{"ambiguous", "Add a bottle of aspirin to Bin 8, using the Medicine type, expiring 03/04.", "", "", false},
		{"query-soon", "What medicine expires soon?", "", "", false},
		{"query-tag", "Which items tagged medicine expire soon?", "", "", false},
		{"whereis", "Where is my Tylenol?", "", "", false},
		{"correct", "Change my Tylenol expiration date to February 2028.", "2028-02", "month", true},
		{"clear", "Remove the expiration date from my Tylenol.", "", "", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
			defer cancel()
			store := memory.NewStore()
			application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, store, memory.NewAuthorizer())
			providers := liveGoogleVoiceProviders(t, ctx)
			application = application.WithRealtimeVoiceProviders(providers.SpeechToText, providers.ConversationModel, providers.TextToSpeech)
			principal := identity.Principal{ID: "user-1"}
			typeKey, typeName := "medicine", "Medicine"
			if scenario.name == "query-tag" {
				typeKey, typeName = "supplies", "Supplies"
			}
			kind, err := application.CreateInventoryCustomAssetType(ctx, app.CreateCustomAssetTypeInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Key: typeKey, DisplayName: typeName, ExpirationEnabled: true})
			if err != nil {
				t.Fatal(err)
			}
			var tagIDs []string
			if scenario.name == "query-tag" {
				tag, err := application.CreateAssetTag(ctx, app.CreateAssetTagInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Key: "medicine", DisplayName: "Medicine"})
				if err != nil {
					t.Fatal(err)
				}
				tagIDs = []string{tag.ID.String()}
			}
			expires := time.Now().UTC().AddDate(0, 0, 7)
			create := func(title, kindName, parent string, dated bool) string {
				input := app.CreateAssetInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Kind: kindName, Title: title, ParentAssetID: parent, TagIDs: tagIDs}
				if dated {
					input.CustomAssetTypeID = kind.ID.String()
					input.Expiration = &assetapp.ExpirationInput{Date: expires.Format("2006-01-02"), Precision: "day"}
				}
				result, err := application.CreateAssetWithOperation(ctx, input)
				if err != nil {
					t.Fatal(err)
				}
				return result.Asset.ID.String()
			}
			closet := create("Hall closet", "location", "", false)
			bin := create("Bin 8", "container", closet, false)
			tylenol := create("Tylenol", "item", bin, true)
			if scenario.name == "query-tag" {
				upcoming := expires
				expires = time.Now().UTC().AddDate(0, 0, -5)
				create("Expired syrup", "item", bin, true)
				expires = time.Now().UTC().AddDate(0, 0, 120)
				create("Cold tablets", "item", bin, true)
				create("Undated ointment", "item", bin, false)
				expires = upcoming
				tagIDs = nil
				create("Untagged supplies", "item", bin, true)
			}
			list := app.ListAssetsInput{Principal: principal, TenantID: "tenant-home", InventoryID: "inventory-home", Limit: 100}
			before, err := application.ListAssets(ctx, list)
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}}})
			if err != nil {
				t.Fatal(err)
			}
			defer connection.Close(websocket.StatusNormalClosure, "")
			connection.SetReadLimit(maxRealtimeVoiceFrameBytes)
			start := realtimeVoiceStartMessage("tenant-home", "inventory-home")
			start["conversationContinuity"], start["developerDiagnostics"] = true, true
			writeRealtimeMessage(t, ctx, connection, start)
			started := readRealtimeMessage(t, ctx, connection)
			if started["type"] != "session.started" {
				t.Fatalf("start: %+v", started)
			}
			t.Logf("EXPIRATION_INPUT %s", scenario.text)
			writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "text.input", "seq": 2, "sessionId": started["sessionId"], "text": scenario.text})
			turn := readLiveInteractionTurn(t, ctx, connection, time.Now(), "expiration-"+scenario.name, 0)
			if scenario.write {
				if turn.terminal != "action.plan.proposed" {
					t.Fatalf("expected approval review: %+v", turn)
				}
				plan, found, err := store.ActionPlanByID(ctx, "tenant-home", "inventory-home", turn.planID)
				if err != nil || !found {
					t.Fatal("missing proposal", err)
				}
				if len(plan.Commands) != 1 {
					t.Fatalf("unexpected commands: %+v", plan.Commands)
				}
				for _, command := range plan.Commands {
					t.Logf("EXPIRATION_PROPOSAL kind=%s arguments=%s", command.Kind, command.ArgumentsJSON)
					var args map[string]any
					if err := json.Unmarshal([]byte(command.ArgumentsJSON), &args); err != nil {
						t.Fatal(err)
					}
					if scenario.name == "correct" || scenario.name == "clear" {
						if string(command.Kind) != "update_asset" || args["assetId"] != tylenol {
							t.Fatalf("wrong correction operation or item: %s %+v", command.Kind, args)
						}
					} else if string(command.Kind) != "create_asset" || args["customAssetTypeId"] != kind.ID.String() || args["parentAssetId"] != bin {
						t.Fatalf("wrong create type or placement: %s %+v", command.Kind, args)
					}
					date, present := args["expiration"]
					if !present {
						t.Fatal("requested expiration omitted")
					}
					if scenario.name == "clear" {
						if date != nil {
							t.Fatal("removal lost")
						}
					} else {
						value, ok := date.(map[string]any)
						if !ok || value["date"] != scenario.date || value["precision"] != scenario.precision {
							t.Fatalf("wrong date: %+v", date)
						}
					}
				}
			} else if turn.terminal != "session.completed" {
				t.Fatalf("expected answer or clarification: %+v", turn)
			}
			if !scenario.write {
				raw, err := os.ReadFile(filepath.Join(os.Getenv("STUFF_STASH_VOICE_EVIDENCE_DIR"), "expiration-"+scenario.name, "first-events.json"))
				if err != nil {
					t.Fatal(err)
				}
				var events []map[string]any
				if err := json.Unmarshal(raw, &events); err != nil {
					t.Fatal(err)
				}
				response := findRealtimeEvent(t, events, "assistant.response.completed")["response"].(map[string]any)
				display, _ := response["displayResponse"].(string)
				if strings.TrimSpace(display) == "" {
					t.Fatal("empty answer")
				}
				if scenario.name == "ambiguous" {
					if !strings.Contains(display, "?") {
						t.Fatalf("expected date clarification: %s", display)
					}
				} else {
					artifacts, _ := response["artifacts"].([]any)
					matched := false
					for _, raw := range artifacts {
						item, _ := raw.(map[string]any)
						if scenario.name == "query-tag" && item["type"] == "asset_reference" && item["assetId"] != tylenol {
							t.Fatalf("tagged upcoming answer included an ineligible item: %+v", item)
						}
						if item["assetId"] == tylenol {
							matched = true
						}
					}
					lower := strings.ToLower(display)
					if !strings.Contains(lower, "expir") || !containsLiveExpirationDate(display, expires) {
						t.Fatalf("answer lost exact expiration date: %s", display)
					}
					if !matched {
						t.Fatalf("answer omitted matching Tylenol reference: %+v", response)
					}
				}
			}
			if turn.audioBytes != 0 {
				t.Fatal("typed input unexpectedly spoke")
			}
			after, err := application.ListAssets(ctx, list)
			if err != nil || !reflect.DeepEqual(before.Items, after.Items) {
				t.Fatal("unapproved conversation mutated inventory", err)
			}
		})
	}
}

func containsLiveExpirationDate(display string, expires time.Time) bool {
	month := "(?:" + expires.Format("January") + "|" + expires.Format("Jan") + "[.]?)"
	day := strconv.Itoa(expires.Day()) + "(?:st|nd|rd|th)?"
	year := expires.Format("2006")
	pattern := `(?i)\b(?:` + expires.Format("2006-01-02") + "|" + month + `\s+` + day + `,?\s+` + year + "|" + day + `\s+` + month + `,?\s+` + year + `)\b`
	return regexp.MustCompile(pattern).MatchString(display)
}
func TestLiveExpirationDateAssertionRejectsMonthOnly(t *testing.T) {
	date := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	if containsLiveExpirationDate("Expires September 2026", date) {
		t.Fatal("year mistaken for day")
	}
	for _, text := range []string{"Expires September 20, 2026", "Expires 20 September 2026", "Expires 2026-09-20"} {
		if !containsLiveExpirationDate(text, date) {
			t.Fatal("exact date rejected", text)
		}
	}
}
