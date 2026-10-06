package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"text/tabwriter"
)

func (o Output) invitation(v ports.Invitation) error {
	fields := [][2]string{{"Invitation", v.ID}, {"Email", v.Email}, {"Relationship", string(v.Relationship)}, {"Status", v.Status}, {"Expires", v.ExpiresAt}, {"Expired", strconv.FormatBool(v.IsExpired)}, {"Inviter", v.InviterPrincipalID}, {"Household", v.TenantID}, {"Inventory", v.InventoryID}}
	if v.AcceptedPrincipalID != nil {
		fields = append(fields, [2]string{"Accepted principal", *v.AcceptedPrincipalID})
	}
	return o.details(fields)
}
func (o Output) invitations(r ports.Result[[]ports.Invitation]) error {
	w := tabwriter.NewWriter(o.Stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "ID\tEMAIL\tROLE\tSTATUS\tEXPIRES\tEXPIRED"); err != nil {
		return err
	}
	for _, v := range r.Data {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\n", strconv.Quote(v.ID), strconv.Quote(v.Email), strconv.Quote(string(v.Relationship)), strconv.Quote(v.Status), strconv.Quote(v.ExpiresAt), v.IsExpired); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return o.pagination(r.Pagination)
}

func (o Output) invitationPreview(v ports.InvitationPreview) error {
	return o.details([][2]string{{"Inventory", v.InventoryName}, {"Inventory ID", v.InventoryID}, {"Role", string(v.Relationship)}, {"Status", v.Status}, {"Expires", v.ExpiresAt}, {"Expired", strconv.FormatBool(v.IsExpired)}})
}
