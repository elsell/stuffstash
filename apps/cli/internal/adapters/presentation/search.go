package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
	"text/tabwriter"
)

func (o Output) searchResults(result ports.Result[[]ports.SearchResult]) error {
	if len(result.Data) == 0 {
		if _, err := fmt.Fprintln(o.Stdout, "No matching assets."); err != nil {
			return err
		}
		return o.pagination(result.Pagination)
	}
	w := tabwriter.NewWriter(o.Stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "ASSET\tINVENTORY\tLOCATION\tMATCH\tID"); err != nil {
		return err
	}
	for _, item := range result.Data {
		path := make([]string, 0, len(item.AncestorPath))
		for _, parent := range item.AncestorPath {
			path = append(path, parent.Title)
		}
		matches := make([]string, 0, len(item.Matches))
		for _, match := range item.Matches {
			matches = append(matches, match.Field+": "+match.Value)
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", strconv.Quote(item.Asset.Title), strconv.Quote(item.Inventory.Name), strconv.Quote(strings.Join(path, " / ")), strconv.Quote(strings.Join(matches, ", ")), strconv.Quote(item.Asset.ID)); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return o.pagination(result.Pagination)
}
