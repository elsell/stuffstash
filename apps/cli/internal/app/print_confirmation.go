package app

import (
	"context"
	"strconv"
)

func (r Runner) confirmPrintCancellation(ctx context.Context, o Options) error {
	if len(o.Command) != 3 || o.Command[0] != "print-jobs" || o.Command[1] != "cancel" {
		return nil
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Print job: " + strconv.Quote(o.Command[2])); err != nil {
		return err
	}
	if err := r.Output.Notice("A label might already have printed. Examine the job before printing again."); err != nil {
		return err
	}
	return r.confirmAction(ctx, o, "Cancel print job", "Cancel job", "Request cancellation. A label might already have printed.")
}
