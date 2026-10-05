package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isAssetWrite(o Options) bool {
	return len(o.Command) > 1 && o.Command[0] == "assets" && (o.Command[1] == "create" || o.Command[1] == "update")
}
func (r Runner) prepareAssetInput(ctx context.Context, o Options) (Options, error) {
	if o.Title == "" {
		if r.TextInput == nil || o.JSON || o.NoInput {
			return o, ports.Failure("usage", "Supply --title TITLE or --input FILE for this command.")
		}
		title, err := r.TextInput.ReadText(ctx, "Asset title", 160)
		if err != nil {
			return o, err
		}
		o.Title = title
	}
	if o.Command[1] == "create" {
		if o.Kind == "" {
			if r.Picker == nil || o.JSON || o.NoInput {
				return o, ports.Failure("usage", "Supply --kind item, container, or location, or use --input FILE.")
			}
			kind, err := r.Picker.Pick(ctx, "Asset kind", []ports.Choice{{ID: "item", Label: "Item", Detail: "A belonging"}, {ID: "container", Label: "Container", Detail: "A box, bin, or shelf"}, {ID: "location", Label: "Location", Detail: "A room or place"}})
			if err != nil {
				return o, err
			}
			o.Kind = kind
		}
		if o.Kind != "item" && o.Kind != "container" && o.Kind != "location" {
			return o, ports.Failure("usage", "Use --kind item, container, or location.")
		}
	}
	return o, nil
}
