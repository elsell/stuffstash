package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (r Runner) confirmAction(ctx context.Context, o Options, title, action, detail string) error {
	if o.Yes {
		return nil
	}
	if r.Picker == nil || o.JSON || o.NoInput {
		return ports.Failure("usage", "Examine the target. Add --yes to approve this action.")
	}
	selected, err := r.Picker.Pick(ctx, title, []ports.Choice{{ID: "cancel", Label: "Cancel", Detail: "Keep the resource unchanged"}, {ID: "confirm", Label: action, Detail: detail}})
	if err != nil {
		return err
	}
	if selected != "confirm" {
		return context.Canceled
	}
	return nil
}
