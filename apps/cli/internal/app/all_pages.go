package app

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// SupportsAllPages is shared by execution, help and completion. It excludes
// read-like mutations and lists whose API does not accept a cursor.
func SupportsAllPages(command []string) bool {
	if len(command) < 2 {
		return false
	}
	switch strings.Join(command[:2], " ") {
	case "tenants list", "tenants audit", "inventories list", "inventories audit", "assets list", "assets search", "assets expiration", "assets checked-out", "tags list", "access-grants list", "invitations list", "notifications list", "archive-jobs list", "asset-types list", "field-definitions list", "workflows list", "printers list", "print-jobs list":
		return len(command) == 2
	case "assets checkouts", "assets activity", "attachments list":
		return len(command) == 3 && command[2] != ""
	case "workflows revisions":
		return len(command) == 4 && command[2] == "list" && command[3] != ""
	case "evaluation cases", "evaluation runs":
		return len(command) == 3 && command[2] == "list"
	case "evaluation revisions":
		return len(command) == 4 && command[2] == "list" && command[3] != ""
	case "connectors print":
		return len(command) == 3 && command[2] == "list" || len(command) == 4 && command[2] == "attempts" && command[3] == "list"
	}
	return false
}

func listPages[T any](ctx context.Context, o Options, fetch func(ports.Page) (ports.Result[[]T], error)) (ports.Result[[]T], error) {
	return collectPages(ctx, o, fetch, func(prior, next []T) []T { return append(prior, next...) })
}

// collectPages changes only the request cursor. It returns no partial data on
// failure, and uses the final response metadata rather than inventing a snapshot.
func collectPages[T any](ctx context.Context, o Options, fetch func(ports.Page) (ports.Result[T], error), merge func(T, T) T) (ports.Result[T], error) {
	if !o.AllPages {
		return fetch(o.Page)
	}
	var zero, result ports.Result[T]
	page := o.Page
	seen := map[string]bool{page.Cursor: true}
	first := true
	for {
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		next, err := fetch(page)
		if canceled := ctx.Err(); canceled != nil {
			return zero, canceled
		}
		if err != nil {
			return zero, err
		}
		if next.Pagination == nil {
			return zero, paginationFailure()
		}
		if !first {
			next.Data = merge(result.Data, next.Data)
		}
		result = next
		first = false
		if !next.Pagination.HasMore {
			return result, nil
		}
		cursor := next.Pagination.NextCursor
		if cursor == nil || *cursor == "" || seen[*cursor] {
			return zero, paginationFailure()
		}
		seen[*cursor] = true
		page.Cursor = *cursor
	}
}
func paginationFailure() error {
	return ports.Failure("protocol", "The server did not provide a usable next page. No complete list was returned. Use --limit and --cursor to examine individual pages, or contact the server administrator.")
}
