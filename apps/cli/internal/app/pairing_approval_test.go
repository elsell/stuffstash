package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/presentation"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPairingInputBeforeCredentials(t *testing.T) {
	for _, command := range [][]string{{"connectors", "print", "pairings", "review", "pairing"}, {"connectors", "print", "pairings", "approve", "pairing"}, {"connectors", "print", "rotations", "approve", "connector"}} {
		for _, body := range []string{`{}`, `{"userCode":7}`, `{"userCode":"code","tenantId":"home","inventoryId":"garage","unknown":"secret"}`, `{"userCode":"code","tenantId":"other","inventoryId":"garage"}`, `{"userCode":"code","tenantId":"home","inventoryId":"garage","bindings":[]}`, `{"userCode":"code","tenantId":"home","inventoryId":"garage","bindings":[{"candidateId":"c","printerId":"p"},{"candidateId":"c","printerId":"q"}]}`, `{"pairingId":"pairing","userCode":"code","generation":18446744073709551616}`, `{"pairingId":"pairing","userCode":"code","generation":1.5}`} {
			args := append(append([]string{}, command...), "--input", "-", "--tenant", "home", "--inventory", "garage", "--json")
			if command[3] == "approve" {
				args = append(args, "--yes")
			}
			o, err := Parse(args, func(string) string { return "" })
			if err != nil {
				t.Fatal(err)
			}
			err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o)
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != "usage" {
				t.Fatalf("malformed input reached credentials: %v", err)
			}
		}
		o, err := Parse(command, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		if err = (Runner{}).Run(context.Background(), o); err == nil {
			t.Fatal("missing input accepted")
		}
	}
}

type pairingCodeInput struct{}

func (pairingCodeInput) ReadSecret(context.Context, string, int) (string, error) {
	return "hidden-code", nil
}

type pairingBindingPicker struct {
	t       *testing.T
	notice  *bytes.Buffer
	confirm bool
}

func (p pairingBindingPicker) Pick(_ context.Context, title string, choices []ports.Choice) (string, error) {
	if strings.HasPrefix(title, "Bind candidate") {
		for _, choice := range choices {
			if choice.ID == "printer:incompatible" {
				p.t.Fatal("incompatible printer offered")
			}
		}
		if !strings.Contains(p.notice.String(), "fingerprint") {
			p.t.Fatal("binding before review display")
		}
		if strings.Contains(title, `"candidate"`) {
			for _, choice := range choices {
				if choice.ID == "printer:second" {
					return choice.ID, nil
				}
			}
			p.t.Fatal("later printer page absent")
		}
		return "skip", nil
	}
	for _, fragment := range []string{`household: "home"`, `inventory: "garage"`, `Bind candidate "candidate" to printer "second"`} {
		if !strings.Contains(p.notice.String(), fragment) {
			p.t.Fatalf("missing displayed binding/scope: %s", p.notice)
		}
	}
	if choices[0].ID != "cancel" {
		p.t.Fatal("confirmation not cancel-first")
	}
	if p.confirm {
		return "confirm", nil
	}
	return "cancel", nil
}
func TestPairingInteractiveBindingsAndDecline(t *testing.T) {
	mutations, lists := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/print-connector-pairings/pairing/review":
			raw, _ := io.ReadAll(r.Body)
			for _, v := range []string{`"userCode":"hidden-code"`, `"tenantId":"home"`, `"inventoryId":"garage"`} {
				if !strings.Contains(string(raw), v) {
					t.Errorf("review mismatch %s", raw)
				}
			}
			io.WriteString(w, `{"data":{"id":"pairing","name":"Kitchen","publicKeyFingerprint":"fingerprint","rotation":false,"candidates":[{"id":"candidate","name":"same name","adapterId":"a"},{"id":"skipped","name":"same name","adapterId":"a"}]}}`)
		case "/tenants/home/inventories/garage/printers":
			lists++
			if r.URL.Query().Get("cursor") == "" {
				io.WriteString(w, `{"data":[{"id":"incompatible","name":"same name","adapterId":"wrong"}],"meta":{"pagination":{"hasMore":true,"nextCursor":"after"}}}`)
			} else {
				io.WriteString(w, `{"data":[{"id":"second","name":"same name","adapterId":"a"}],"meta":{"pagination":{"hasMore":false,"nextCursor":null}}}`)
			}
		case "/print-connector-pairings/pairing/approval":
			mutations++
			raw, _ := io.ReadAll(r.Body)
			var v pairingApprovalInput
			if err := json.Unmarshal(raw, &v); err != nil {
				t.Fatal(err)
			}
			if v.UserCode != "hidden-code" || v.TenantID != "home" || v.InventoryID != "garage" || len(v.Bindings) != 1 || v.Bindings[0].CandidateID != "candidate" || v.Bindings[0].PrinterID != "second" {
				t.Errorf("wrong exact bindings %s", raw)
			}
			io.WriteString(w, `{"data":{"id":"connector","generation":1}}`)
		default:
			t.Errorf("unexpected API %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	api, err := httpapi.New(server.URL, "human", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	for _, confirm := range []bool{false, true} {
		var out, notice bytes.Buffer
		runner := Runner{SecretInput: pairingCodeInput{}, Picker: pairingBindingPicker{t, &notice, confirm}, Output: presentation.Output{Stdout: &out, Stderr: &notice}, Observer: lifecycleObserver{}, PairingApprovalAPI: func(string, string) (ports.PairingApprovalAPI, error) { return api, nil }, PrintingAPI: func(string, string) (ports.HumanPrintingAPI, error) { return api, nil }}
		o := Options{Server: server.URL, Scope: ports.Scope{Tenant: "home", Inventory: "garage"}, Command: []string{"connectors", "print", "pairings", "approve", "pairing"}}
		o, err = runner.preparePairingApproval(context.Background(), o)
		if err != nil {
			t.Fatal(err)
		}
		err = runner.pairingApprovalCommand(context.Background(), o, "human")
		if !confirm && (!errors.Is(err, context.Canceled) || mutations != 0) {
			t.Fatalf("declined approval mutated %v", err)
		}
		if confirm && (err != nil || mutations != 1) {
			t.Fatalf("approval failed %v", err)
		}
		if strings.Contains(out.String()+notice.String(), "hidden-code") {
			t.Fatal("code exposed")
		}
	}
	if lists != 4 {
		t.Fatalf("printer paging %d", lists)
	}
}
