package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"text/tabwriter"
)

func (o Output) accessGrant(v ports.AccessGrant) error {
	return o.details([][2]string{{"Principal ID", v.PrincipalID}, {"Relationship", string(v.Relationship)}, {"Household ID", v.TenantID}, {"Inventory ID", v.InventoryID}})
}
func (o Output) accessGrants(r ports.Result[[]ports.AccessGrant]) error {
	w := tabwriter.NewWriter(o.Stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "PRINCIPAL\tRELATIONSHIP\tHOUSEHOLD\tINVENTORY"); err != nil {
		return err
	}
	for _, v := range r.Data {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", strconv.Quote(v.PrincipalID), strconv.Quote(string(v.Relationship)), strconv.Quote(v.TenantID), strconv.Quote(v.InventoryID)); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return o.pagination(r.Pagination)
}
