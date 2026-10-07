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

func TestVoiceProviderUpdateInvalidBeforeCredentials(t *testing.T) {
	for _, body := range []string{`{"speechToTextProfileId":false}`, `{"languageInferenceProfileId":42}`, `{"textToSpeechProfileId":[]}`, `{"extra":"private"}`} {
		o, err := Parse([]string{"voice-provider", "update", "--input", "-", "--yes", "--json"}, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o)
		var failure *ports.Error
		if !errors.As(err, &failure) || failure.Category != "usage" {
			t.Fatalf("invalid input reached auth: %v", err)
		}
	}
	for _, args := range [][]string{{"voice-provider", "update", "--yes", "--no-input"}, {"voice-provider", "update", "--json"}} {
		o, err := Parse(args, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		if err = (Runner{}).Run(context.Background(), o); err == nil {
			t.Fatal("missing scripted input accepted")
		}
	}
}

type voiceSelectionPicker struct {
	t       *testing.T
	notice  *bytes.Buffer
	cancel  bool
	prompts int
}

func (p *voiceSelectionPicker) Pick(_ context.Context, title string, choices []ports.Choice) (string, error) {
	p.prompts++
	for _, c := range choices {
		if strings.Contains(c.Label, "Archived") || strings.Contains(c.Label, "Wrong capability") {
			p.t.Fatal("ineligible profile offered")
		}
	}
	if title == "Replace voice provider selections" {
		for _, part := range []string{`Household: "home"`, `explicit "old"`, "automatic (server selects a profile)", "Replace all three", "no version check"} {
			if !strings.Contains(p.notice.String(), part) {
				p.t.Fatalf("missing review %s in %s", part, p.notice)
			}
		}
		if p.cancel {
			return "cancel", nil
		}
		return "confirm", nil
	}
	if choices[0].ID != "cancel" {
		p.t.Fatal("picker must offer safe cancellation")
	}
	if title == "Spoken output" {
		for _, c := range choices {
			if strings.Contains(c.Label, `"disabled"`) {
				if !strings.Contains(c.Detail, "disabled") {
					p.t.Fatal("disabled state hidden")
				}
				return c.ID, nil
			}
		}
		p.t.Fatal("matching disabled profile unavailable")
	}
	return "keep", nil
}
func TestVoiceProviderUpdateGuidedKeepsExplicitAndAutomatic(t *testing.T) {
	for _, scenario := range []struct{ cancel, outer bool }{{false, true}, {false, false}, {true, false}} {
		cancel := scenario.cancel
		puts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer owner" || !strings.HasPrefix(r.URL.Path, "/tenants/home/") {
				t.Fatal("wrong guided auth or scope")
			}
			if r.Method == "PUT" {
				puts++
				var body map[string]*string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body["speechToTextProfileId"] == nil || *body["speechToTextProfileId"] != "old" || body["languageInferenceProfileId"] != nil || body["textToSpeechProfileId"] == nil || *body["textToSpeechProfileId"] != "disabled" {
					t.Fatalf("changed guided choices: %+v", body)
				}
				io.WriteString(w, `{"data":null,"meta":{}}`)
				return
			}
			if strings.HasSuffix(r.URL.Path, "provider-profiles") {
				io.WriteString(w, `{"data":[{"id":"old","displayName":"Existing","capability":"speech_to_text","lifecycleState":"enabled"},{"id":"archived","displayName":"Archived","capability":"text_to_speech","lifecycleState":"archived"},{"id":"wrong","displayName":"Wrong capability","capability":"other","lifecycleState":"enabled"},{"id":"disabled","displayName":"Spoken output","capability":"text_to_speech","lifecycleState":"disabled"}],"meta":{}}`)
				return
			}
			// Missing slot ID falls back to the outer explicit ID; implicit effective ID stays automatic.
			if scenario.outer {
				io.WriteString(w, `{"data":{"tenantId":"home","profileIds":{"speechToText":"old","languageInference":"implicit"},"slots":[{"capability":"speech_to_text","selectionSource":"explicit"},{"capability":"language_inference","selectionSource":"implicit","selectedProfileId":"implicit"}]},"meta":{}}`)
			} else {
				io.WriteString(w, `{"data":{"tenantId":"home","profileIds":{},"slots":[{"capability":"speech_to_text","selectionSource":"explicit","selectedProfileId":"old"},{"capability":"language_inference","selectionSource":"implicit","selectedProfileId":"implicit"}]},"meta":{}}`)
			}
		}))
		api, err := httpapi.New(server.URL, "owner", server.Client())
		if err != nil {
			t.Fatal(err)
		}
		var out, notice bytes.Buffer
		picker := &voiceSelectionPicker{t: t, notice: &notice, cancel: cancel}
		runner := Runner{Picker: picker, Output: presentation.Output{Stdout: &out, Stderr: &notice}, Observer: lifecycleObserver{}, ProviderProfilesAPI: func(string, string) (ports.ProviderProfilesAPI, error) { return api, nil }}
		o := Options{Server: server.URL, Scope: ports.Scope{Tenant: "home"}, Command: []string{"voice-provider", "update"}}
		err = runner.updateVoiceProvider(context.Background(), o, "owner", api)
		if cancel {
			if !errors.Is(err, context.Canceled) || puts != 0 {
				t.Fatalf("declined update: %v %d", err, puts)
			}
		} else if err != nil || puts != 1 {
			t.Fatalf("guided update: %v %d", err, puts)
		}
		server.Close()
	}
}

func TestVoiceProviderUpdateCannotKeepHiddenExplicitID(t *testing.T) {
	puts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			puts++
			return
		}
		if strings.HasSuffix(r.URL.Path, "provider-profiles") {
			io.WriteString(w, `{"data":[],"meta":{}}`)
			return
		}
		io.WriteString(w, `{"data":{"profileIds":{},"slots":[{"capability":"speech_to_text","selectionSource":"explicit","readiness":"invalid_selection"}]},"meta":{}}`)
	}))
	defer server.Close()
	api, err := httpapi.New(server.URL, "owner", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	runner := Runner{Picker: firstScopeChoice{}, ProviderProfilesAPI: func(string, string) (ports.ProviderProfilesAPI, error) { return api, nil }}
	_, err = runner.guideVoiceProvider(context.Background(), Options{Scope: ports.Scope{Tenant: "home"}}, "owner", api)
	var failure *ports.Error
	if !errors.As(err, &failure) || failure.Category != "usage" || puts != 0 {
		t.Fatalf("hidden explicit selection cleared: %v", err)
	}
}
