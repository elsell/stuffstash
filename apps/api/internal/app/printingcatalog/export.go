// Package printingcatalog exports deterministic synthetic fixtures using the
// injected production renderer, without API, database, or device access.
package printingcatalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const SchemaVersion = 1
const FixtureURL = "https://example.invalid/l/v1/01JEXAMPLEINSTANCE/01JEXAMPLELABEL"

type Renderer interface {
	ports.LabelRenderer
	ports.LabelDisplayTransformer
}
type MediaCatalog struct {
	SchemaVersion int                      `json:"schema_version"`
	Media         []printing.MediaSnapshot `json:"media"`
}
type Example struct {
	Template            printing.TemplateSelection `json:"template"`
	Media               printing.MediaSnapshot     `json:"media"`
	PNG                 string                     `json:"png"`
	DisplayPNG          string                     `json:"display_png"`
	PDF                 string                     `json:"pdf"`
	WidthPixels         int                        `json:"width_pixels"`
	HeightPixels        int                        `json:"height_pixels"`
	DisplayWidthPixels  int                        `json:"display_width_pixels"`
	DisplayHeightPixels int                        `json:"display_height_pixels"`
}
type Catalog struct {
	SchemaVersion int                           `json:"schema_version"`
	Notice        string                        `json:"notice"`
	Templates     []printing.TemplateDescriptor `json:"templates"`
	Media         []printing.MediaSnapshot      `json:"media"`
	Examples      []Example                     `json:"examples"`
}
type Bundle struct {
	Catalog Catalog
	Files   map[string][]byte
}

var safeID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,79}$`)

func Export(ctx context.Context, renderer Renderer, templates ports.LabelTemplateCatalog, media []printing.MediaSnapshot) (Bundle, error) {
	catalog := Catalog{SchemaVersion: SchemaVersion, Notice: "Generated from executable template registrations. Fixtures do not establish printer support or physical verification.", Templates: templates.Templates(), Media: append([]printing.MediaSnapshot{}, media...), Examples: []Example{}}
	sort.Slice(catalog.Templates, func(i, j int) bool {
		a, b := catalog.Templates[i], catalog.Templates[j]
		if a.ID == b.ID {
			return a.Version < b.Version
		}
		return a.ID < b.ID
	})
	sort.Slice(catalog.Media, func(i, j int) bool {
		a, b := catalog.Media[i], catalog.Media[j]
		if a.PresetID == b.PresetID {
			return a.Version < b.Version
		}
		return a.PresetID < b.PresetID
	})
	seen := map[string]bool{}
	for _, template := range catalog.Templates {
		key := fmt.Sprintf("%s-v%d", template.ID, template.Version)
		if !safeID.MatchString(string(template.ID)) || template.Version == 0 || seen[key] {
			return Bundle{}, errors.New("invalid or duplicate template identity")
		}
		seen[key] = true
	}
	seen = map[string]bool{}
	for _, m := range catalog.Media {
		key := fmt.Sprintf("%s-v%d", m.PresetID, m.Version)
		if !safeID.MatchString(m.PresetID) || m.Version == 0 || seen[key] {
			return Bundle{}, errors.New("invalid or duplicate media identity")
		}
		seen[key] = true
	}
	bundle := Bundle{Files: map[string][]byte{}}
	for _, m := range catalog.Media {
		for _, template := range catalog.Templates {
			for _, showReference := range []bool{true, false} {
				selection := printing.TemplateSelection{ID: template.ID, Version: template.Version, Options: printing.TemplateOptions{ShowReference: showReference}}
				prefix := fmt.Sprintf("%s-v%d-%s-v%d-reference-%t", template.ID, template.Version, m.PresetID, m.Version, showReference)
				req := printing.RenderRequest{Content: printing.ContentSnapshot{QRURL: FixtureURL, Title: "Garage tools", Reference: "STASH-012345"}, Template: selection, Media: m, Format: printing.FormatPNG}
				rendered, err := renderer.Render(ctx, req)
				if err != nil {
					return Bundle{}, fmt.Errorf("render fixture %s: %w", prefix, err)
				}
				display, err := renderer.DisplayPNG(rendered)
				if err != nil {
					return Bundle{}, fmt.Errorf("orient fixture %s: %w", prefix, err)
				}
				req.Format = printing.FormatPDF
				pdf, err := renderer.Render(ctx, req)
				if err != nil {
					return Bundle{}, fmt.Errorf("render PDF fixture %s: %w", prefix, err)
				}
				example := Example{Template: selection, Media: m, PNG: prefix + ".png", DisplayPNG: prefix + "-display.png", PDF: prefix + ".pdf", WidthPixels: rendered.WidthPixels, HeightPixels: rendered.HeightPixels, DisplayWidthPixels: rendered.WidthPixels, DisplayHeightPixels: rendered.HeightPixels}
				if m.DisplayRotation == 90 || m.DisplayRotation == 270 {
					example.DisplayWidthPixels, example.DisplayHeightPixels = example.DisplayHeightPixels, example.DisplayWidthPixels
				}
				bundle.Files[example.PNG] = rendered.Content
				bundle.Files[example.DisplayPNG] = display
				bundle.Files[example.PDF] = pdf.Content
				catalog.Examples = append(catalog.Examples, example)
			}
		}
	}
	bundle.Catalog = catalog
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return Bundle{}, err
	}
	bundle.Files["catalog.json"] = append(data, '\n')
	names := make([]string, 0, len(bundle.Files))
	for name := range bundle.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	manifest, err := json.MarshalIndent(struct {
		SchemaVersion int      `json:"schema_version"`
		Files         []string `json:"files"`
	}{SchemaVersion, names}, "", "  ")
	if err != nil {
		return Bundle{}, err
	}
	bundle.Files["manifest.json"] = append(manifest, '\n')
	return bundle, nil
}
