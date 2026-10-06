package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/labels"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type labelCatalogFake struct {
	defaults ports.InventoryPrintDefaults
	media    []printing.Media
	rendered []ports.LabelRenderSelection
	printer  string
}

func (f *labelCatalogFake) PrintDefaults(context.Context, ports.Scope) (ports.InventoryPrintDefaults, error) {
	return f.defaults, nil
}
func (f *labelCatalogFake) LabelMedia(_ context.Context, _ ports.Scope, printer string) ([]printing.Media, error) {
	f.printer = printer
	return f.media, nil
}
func (*labelCatalogFake) LabelTemplates(context.Context, ports.Scope) (ports.Result[[]ports.LabelTemplate], error) {
	return ports.Result[[]ports.LabelTemplate]{}, nil
}
func (*labelCatalogFake) ResolveLabel(context.Context, labels.Reference) (ports.Result[ports.ResolvedLabel], error) {
	return ports.Result[ports.ResolvedLabel]{}, nil
}
func (f *labelCatalogFake) RenderLabel(_ context.Context, _ ports.Scope, _ string, s ports.LabelRenderSelection) (ports.LabelArtifact, error) {
	f.rendered = append(f.rendered, s)
	return ports.LabelArtifact{Content: []byte("verified"), Format: s.Format, SHA256: "digest"}, nil
}

type labelFilesFake struct{ files map[string]string }

func (f *labelFilesFake) Publish(_ context.Context, path string, content []byte) error {
	if _, ok := f.files[path]; ok {
		return ports.Failure("file", "exists")
	}
	f.files[path] = string(content)
	return nil
}

func TestStandaloneLabelUsesCatalogGeometryAndIndependentTemplateDefaults(t *testing.T) {
	media := printing.Media{PresetID: "brother-ql800-29x90", Version: ^uint32(0), WidthMicrometers: 29000, HeightMicrometers: 89800, RasterWidth: 306, RasterHeight: 991}
	fake := &labelCatalogFake{defaults: ports.InventoryPrintDefaults{PrinterID: "garage", TemplateID: "qr-title", TemplateVersion: ^uint32(0), ShowReference: true}, media: []printing.Media{media}}
	files := &labelFilesFake{files: map[string]string{}}
	options := Options{Command: []string{"labels", "render", "tool"}, Scope: ports.Scope{Tenant: "home", Inventory: "garage"}, Format: "png", OutputPath: "label.png"}
	if err := validateCommand(options); err != nil {
		t.Fatal(err)
	}
	if _, err := executeLabels(context.Background(), fake, files, options); err != nil {
		t.Fatal(err)
	}
	if fake.printer != "garage" || len(fake.rendered) != 1 || fake.rendered[0].Media != media || !fake.rendered[0].ShowReference {
		t.Fatalf("wrong destination/defaults: %+v", fake)
	}
	options.OutputPath = "standalone.pdf"
	options.Format = "pdf"
	options.MediaPreset = media.PresetID
	options.TemplateID = "qr-only"
	options.ShowReferenceSet = true
	options.ShowReference = false
	if _, err := executeLabels(context.Background(), fake, files, options); err != nil {
		t.Fatal(err)
	}
	if fake.printer != "" || fake.rendered[1].TemplateID != "qr-only" || fake.rendered[1].ShowReference {
		t.Fatal("explicit media/template not honored")
	}
	options.MediaPreset = ""
	options.WidthMM = 29
	options.HeightMM = 50
	options.OutputPath = "unsupported.png"
	if _, err := executeLabels(context.Background(), fake, files, options); err == nil {
		t.Fatal("unsupported dimensions rendered")
	}
	if len(fake.rendered) != 2 || len(files.files) != 2 {
		t.Fatal("invalid media produced output")
	}
	options.HeightMM = 89.8
	if _, err := executeLabels(context.Background(), fake, files, options); err != nil {
		t.Fatal(err)
	}
	options.PrinterID = "garage"
	if err := validateCommand(options); err == nil {
		t.Fatal("conflicting media selectors accepted")
	}
}

func (*labelCatalogFake) AssetLabel(context.Context, ports.Scope, string) (ports.Result[ports.ResolvedLabel], error) {
	return ports.Result[ports.ResolvedLabel]{}, ports.Failure("not_found", "Asset not found.")
}
func (*labelCatalogFake) AssignLabel(context.Context, ports.Scope, string) (ports.Result[ports.ResolvedLabel], error) {
	return ports.Result[ports.ResolvedLabel]{}, ports.Failure("not_found", "Asset not found.")
}
