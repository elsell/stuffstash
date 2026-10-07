// Package scopeselection resolves missing scope from authorized API catalogs.
package scopeselection

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func Choose(ctx context.Context, api ports.ScopeCatalog, picker ports.Selector, scope ports.Scope, inventoryRequired bool) (ports.Scope, error) {
	if scope.Tenant == "" {
		choices, err := catalogChoices(ctx, func(p ports.Page) ([]ports.Choice, *ports.Pagination, error) {
			result, err := api.Tenants(ctx, p)
			if err != nil {
				return nil, nil, err
			}
			items := make([]ports.Choice, 0, len(result.Data))
			for _, entry := range result.Data {
				items = append(items, ports.Choice{ID: entry.ID, Label: entry.Name, Detail: entry.ID + " · " + entry.Lifecycle})
			}
			return items, result.Pagination, nil
		})
		if err != nil {
			return scope, err
		}
		selected, err := pick(ctx, picker, "Choose a household", choices)
		if err != nil {
			return scope, err
		}
		scope.Tenant = selected
	}
	if inventoryRequired && scope.Inventory == "" {
		choices, err := catalogChoices(ctx, func(p ports.Page) ([]ports.Choice, *ports.Pagination, error) {
			result, err := api.Inventories(ctx, scope, p)
			if err != nil {
				return nil, nil, err
			}
			items := make([]ports.Choice, 0, len(result.Data))
			for _, entry := range result.Data {
				items = append(items, ports.Choice{ID: entry.ID, Label: entry.Name, Detail: entry.ID + " · " + entry.Lifecycle})
			}
			return items, result.Pagination, nil
		})
		if err != nil {
			return scope, err
		}
		scope.Inventory, err = pick(ctx, picker, "Choose an inventory", choices)
		if err != nil {
			return scope, err
		}
	}
	return scope, nil
}
func catalogChoices(ctx context.Context, fetch func(ports.Page) ([]ports.Choice, *ports.Pagination, error)) ([]ports.Choice, error) {
	var choices []ports.Choice
	request := ports.Page{Limit: 100}
	seen := map[string]bool{"": true}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		items, page, err := fetch(request)
		if err != nil {
			return nil, err
		}
		choices = append(choices, items...)
		if page == nil || !page.HasMore {
			return choices, nil
		}
		if page.NextCursor == nil || *page.NextCursor == "" || seen[*page.NextCursor] {
			return nil, ports.Failure("protocol", "The server returned an incorrect page cursor. Try again or update the server.")
		}
		request.Cursor = *page.NextCursor
		seen[request.Cursor] = true
	}
}
func pick(ctx context.Context, picker ports.Selector, title string, choices []ports.Choice) (string, error) {
	if len(choices) == 0 {
		return "", ports.Failure("not_found", "No authorized choices are available. Ask an administrator for access or create a resource first.")
	}
	selected, err := picker.Pick(ctx, title, choices)
	if err != nil {
		return "", err
	}
	for _, choice := range choices {
		if choice.ID != "" && choice.ID == selected {
			return selected, nil
		}
	}
	return "", ports.Failure("validation", "The selection is not in the available choices. Try again.")
}
