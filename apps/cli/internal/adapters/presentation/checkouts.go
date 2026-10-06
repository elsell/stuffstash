package presentation

import "github.com/stuffstash/stuff-stash/cli/internal/ports"

func (o Output) checkoutDetails(c ports.Checkout) error {
	fields := [][2]string{{"Checkout ID", c.ID}, {"Asset ID", c.AssetID}, {"State", c.State}, {"Borrower ID", c.CheckedOutByPrincipalID}, {"Checked out", c.CheckedOutAt}}
	for _, v := range []struct {
		label string
		value *string
	}{{"Checkout notes", c.CheckoutDetails}, {"Returned", c.ReturnedAt}, {"Returned by", c.ReturnedByPrincipalID}, {"Return notes", c.ReturnDetails}, {"Undo ID", c.UndoableOperationID}} {
		if v.value != nil {
			fields = append(fields, [2]string{v.label, *v.value})
		}
	}
	return o.details(fields)
}
