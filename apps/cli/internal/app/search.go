package app

import (
	"context"
	"flag"
	"unicode/utf8"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isSearch(o Options) bool {
	return len(o.Command) > 1 && o.Command[0] == "assets" && o.Command[1] == "search"
}
func validateSearchFlags(o Options, flags *flag.FlagSet) error {
	var unsupported string
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server", "tenant", "inventory", "context", "json", "no-input", "request-id", "color", "help", "query", "mode", "type-id", "tag-id", "lifecycle", "checkout-state", "limit", "cursor", "all-inventories":
		default:
			unsupported = f.Name
		}
	})
	if unsupported != "" {
		return ports.Failure("usage", "Asset search does not accept --"+unsupported+". Remove the option.")
	}
	return validateSearch(o, false)
}
func validateSearch(o Options, scope bool) error {
	if len(o.Command) != 2 {
		return ports.Failure("usage", "Use assets search --query TEXT. Remove extra arguments.")
	}
	q := o.Expiration
	if q.Mode != "" && q.Mode != "exact" && q.Mode != "fuzzy" {
		return ports.Failure("usage", "Use --mode exact or fuzzy.")
	}
	if q.CheckoutState != "" && q.CheckoutState != "any" && q.CheckoutState != "available" && q.CheckoutState != "checked_out" {
		return ports.Failure("usage", "Use --checkout-state any, available, or checked_out.")
	}
	if o.Lifecycle != "" && o.Lifecycle != "active" && o.Lifecycle != "archived" && o.Lifecycle != "all" {
		return ports.Failure("usage", "Use --lifecycle active, archived, or all.")
	}
	if o.Page.Limit < 1 || o.Page.Limit > 100 {
		return ports.Failure("usage", "Use --limit from 1 to 100.")
	}
	if utf8.RuneCountInString(q.Query) > 120 {
		return ports.Failure("usage", "Keep --query to 120 characters or fewer.")
	}
	if utf8.RuneCountInString(q.TypeID) > 128 {
		return ports.Failure("usage", "Keep --type-id to 128 characters or fewer.")
	}
	if scope && missingResourceScope(o) {
		return ports.Failure("usage", "Choose a saved context or supply --tenant and --inventory. Use --all-inventories to search the household.")
	}
	return nil
}
func (r Runner) searchAssets(ctx context.Context, o Options, token string) error {
	if r.SearchAPI == nil {
		return ports.Failure("configuration", "Asset search is not available. Update the CLI and try again.")
	}
	api, err := r.SearchAPI(o.Server, token)
	if err != nil {
		return err
	}
	scope := o.Scope
	if o.AllInventories {
		scope.Inventory = ""
	}
	q := o.Expiration
	result, err := api.SearchAssets(ctx, scope, ports.SearchQuery{Page: o.Page, Query: q.Query, Mode: q.Mode, TypeID: q.TypeID, TagIDs: q.TagIDs, Lifecycle: o.Lifecycle, CheckoutState: q.CheckoutState})
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.assets.search.completed")
	return r.Output.Result(result)
}
