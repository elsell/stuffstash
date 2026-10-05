package app

import (
	"context"
	"strconv"
)

func isAssetLifecycle(o Options) bool {
	return len(o.Command) > 1 && o.Command[0] == "assets" && (o.Command[1] == "archive" || o.Command[1] == "restore" || o.Command[1] == "delete")
}
func (r Runner) confirmAssetLifecycle(ctx context.Context, o Options) error {
	if !isAssetLifecycle(o) {
		return nil
	}
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + "; household: " + strconv.Quote(o.Scope.Tenant) + "; inventory: " + strconv.Quote(o.Scope.Inventory) + "; asset: " + strconv.Quote(o.Command[2])); err != nil {
		return err
	}
	switch o.Command[1] {
	case "archive":
		return r.confirmAction(ctx, o, "Archive asset", "Archive", "Hide this asset from active inventory lists.")
	case "delete":
		return r.confirmAction(ctx, o, "Delete asset", "Delete", "Permanently delete this asset.")
	default:
		return nil
	}
}
