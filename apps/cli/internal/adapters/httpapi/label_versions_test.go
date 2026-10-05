package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLabelSelectionPreservesUnsignedCatalogAndRequestVersions(t *testing.T) {
	fake := &labelServer{}
	captured := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := "/tenants/home/inventories/garage/"
		switch r.URL.Path {
		case prefix + "label-templates":
			io.WriteString(w, `{"data":[{"id":"template","version":4294967295,"defaults":{"show_reference":false}}],"meta":{}}`)
			return
		case prefix + "printer-profiles":
			io.WriteString(w, `{"data":[{"adapterId":"adapter","media":[{"presetId":"preset","version":4294967295}]}],"meta":{}}`)
			return
		case prefix + "printers/printer":
			io.WriteString(w, `{"data":{"id":"printer","revision":1,"media":{"presetId":"preset","version":4294967295}},"meta":{}}`)
			return
		case prefix + "assets/tool/label-renders":
			data, err := io.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
				return
			}
			var request struct {
				Media struct {
					Version uint32 `json:"version"`
				} `json:"media"`
				Template struct {
					Version uint32 `json:"version"`
					Options struct {
						ShowReference bool `json:"show_reference"`
					} `json:"options"`
				} `json:"template"`
			}
			if json.Unmarshal(data, &request) != nil || request.Media.Version != ^uint32(0) || request.Template.Version != ^uint32(0) || request.Template.Options.ShowReference {
				t.Errorf("narrowed flags request: %s", data)
			}
			captured = true
			r.Body = io.NopCloser(bytes.NewReader(data))
		}
		fake.ServeHTTP(w, r)
	}))
	defer server.Close()
	api, err := New(server.URL, "editor", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	scope := ports.Scope{Tenant: "home", Inventory: "garage"}
	templates, err := api.LabelTemplates(context.Background(), scope)
	if err != nil || len(templates.Data) != 1 || templates.Data[0].Version != ^uint32(0) {
		t.Fatalf("templates: %+v %v", templates, err)
	}
	for _, printer := range []string{"", "printer"} {
		media, err := api.LabelMedia(context.Background(), scope, printer)
		if err != nil || len(media) != 1 || media[0].Version != ^uint32(0) {
			t.Fatalf("media: %+v %v", media, err)
		}
		if _, err = api.RenderLabel(context.Background(), scope, "tool", ports.LabelRenderSelection{Media: media[0], TemplateID: "template", TemplateVersion: templates.Data[0].Version, Format: "png"}); err != nil {
			t.Fatal(err)
		}
	}
	if !captured {
		t.Fatal("render was not submitted")
	}
}
