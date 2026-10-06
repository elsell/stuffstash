package app

import (
	"context"
	"strconv"
)

func (r Runner) confirmPrintCancellation(ctx context.Context, o Options) error {
	if len(o.Command) != 3 || o.Command[0] != "print-jobs" || o.Command[1] != "cancel" {
		return nil
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; print job: " + strconv.Quote(o.Command[2])); err != nil {
		return err
	}
	if err := r.Output.Notice("A label might already have printed. Inspect the job before printing again."); err != nil {
		return err
	}
	return r.confirmAction(ctx, o, "Cancel print job", "Cancel job", "Request cancellation. A label might already have printed.")
}
