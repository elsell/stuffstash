package app

import (
	"context"
	"math"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/labels"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isLabelCommand(o Options) bool {
	return len(o.Command) >= 2 && o.Command[0] == "labels" && o.Command[1] != "print"
}
func validateLabelCommand(o Options) error { return validateLabelCommandOptions(o, true) }
func validateLabelCommandOptions(o Options, requireScope bool) error {
	if o.Command[1] == "resolve" && len(o.Command) == 3 {
		if _, err := labels.Parse(o.Command[2]); err != nil {
			return ports.Failure("usage", "invalid or unsupported label link")
		}
		return nil
	}
	if requireScope && (o.Scope.Tenant == "" || o.Scope.Inventory == "") {
		return ports.Failure("usage", "label commands require --tenant and --inventory")
	}
	if (o.Command[1] == "show" || o.Command[1] == "assign") && len(o.Command) == 3 && o.Command[2] != "" {
		if o.IdempotencyKey != "" || o.Page.Cursor != "" || o.Title != "" || o.Kind != "" || o.Parent != "" || o.ConnectorName != "" {
			return ports.Failure("usage", "Label identity commands do not accept input fields or cursors. Remove those options.")
		}
		return nil
	}
	if o.Command[1] == "templates" && len(o.Command) == 2 {
		return nil
	}
	if o.Command[1] != "render" || len(o.Command) != 3 || o.OutputPath == "" || o.OutputPath == "-" || (o.Format != "png" && o.Format != "pdf") {
		return ports.Failure("usage", "use labels render ASSET --format png|pdf --output PATH")
	}
	if o.TemplateVersion > math.MaxInt32 {
		return ports.Failure("usage", "template version exceeds the API limit")
	}
	dimensions := o.WidthMM != 0 || o.HeightMM != 0
	if dimensions && (o.WidthMM <= 0 || o.HeightMM <= 0 || math.IsNaN(o.WidthMM) || math.IsNaN(o.HeightMM) || math.IsInf(o.WidthMM, 0) || math.IsInf(o.HeightMM, 0)) {
		return ports.Failure("usage", "label dimensions must both be positive finite millimeters")
	}
	if (o.PrinterID != "" && (o.MediaPreset != "" || dimensions)) || (o.MediaPreset != "" && dimensions) {
		return ports.Failure("usage", "choose one media selector: --printer, --media-preset, or dimensions")
	}
	return nil
}
func executeLabels(ctx context.Context, api ports.LabelsAPI, files ports.LabelFiles, o Options) (any, error) {
	switch o.Command[1] {
	case "show":
		return api.AssetLabel(ctx, o.Scope, o.Command[2])
	case "assign":
		return api.AssignLabel(ctx, o.Scope, o.Command[2])
	case "templates":
		return api.LabelTemplates(ctx, o.Scope)
	case "resolve":
		ref, err := labels.Parse(o.Command[2])
		if err != nil {
			return nil, ports.Failure("usage", "invalid label link")
		}
		return api.ResolveLabel(ctx, ref)
	}
	defaults, err := api.PrintDefaults(ctx, o.Scope)
	if err != nil {
		return nil, err
	}
	selection := ports.LabelRenderSelection{TemplateID: defaults.TemplateID, TemplateVersion: defaults.TemplateVersion, ShowReference: defaults.ShowReference, Format: o.Format}
	if o.TemplateID != "" {
		selection.TemplateID = o.TemplateID
	}
	if o.TemplateVersion != 0 {
		selection.TemplateVersion = uint32(o.TemplateVersion)
	}
	if o.ShowReferenceSet {
		selection.ShowReference = o.ShowReference
	}
	printer := o.PrinterID
	if printer == "" && o.MediaPreset == "" && o.WidthMM == 0 {
		printer = defaults.PrinterID
		if printer == "" {
			return nil, ports.Failure("usage", "choose --printer, --media-preset, or supported label dimensions")
		}
	}
	media, err := api.LabelMedia(ctx, o.Scope, printer)
	if err != nil {
		return nil, err
	}
	var selected []printing.Media
	for _, m := range media {
		matches := printer != "" || (o.MediaPreset != "" && m.PresetID == o.MediaPreset) || (o.WidthMM > 0 && math.Abs(float64(m.WidthMicrometers)-o.WidthMM*1000) < 0.5 && math.Abs(float64(m.HeightMicrometers)-o.HeightMM*1000) < 0.5)
		if matches {
			duplicate := false
			for _, prior := range selected {
				if prior == m {
					duplicate = true
				}
			}
			if !duplicate {
				selected = append(selected, m)
			}
		}
	}
	if len(selected) != 1 {
		return nil, ports.Failure("usage", "media selection is unsupported or ambiguous; choose an authorized catalog preset")
	}
	selection.Media = selected[0]
	artifact, err := api.RenderLabel(ctx, o.Scope, o.Command[2], selection)
	if err != nil {
		return nil, err
	}
	if files == nil {
		return nil, ports.Failure("configuration", "label file output is unavailable")
	}
	if err = files.Publish(ctx, o.OutputPath, artifact.Content); err != nil {
		return nil, err
	}
	return ports.LabelFileResult{Path: o.OutputPath, Format: artifact.Format, SHA256: artifact.SHA256}, nil
}
