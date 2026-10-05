package presentation

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"strings"
)

func (o Output) assetDetails(a ports.Asset) error {
	fields := [][2]string{{"Title", a.Title}, {"ID", a.ID}, {"Household ID", a.TenantID}, {"Inventory ID", a.InventoryID}, {"Kind", a.Kind}, {"State", a.Lifecycle}}
	if a.Description != "" {
		fields = append(fields, [2]string{"Description", a.Description})
	}
	if a.Parent != nil {
		fields = append(fields, [2]string{"Parent ID", *a.Parent})
	}
	if a.Expiration != nil {
		fields = append(fields, [2]string{"Expires", a.Expiration.Date})
	}
	if a.ExpirationContext != nil {
		fields = append(fields, [2]string{"Expiry state", a.ExpirationContext.State})
	}
	if a.CurrentCheckout != nil {
		fields = append(fields, [2]string{"Checkout", a.CurrentCheckout.State}, [2]string{"Checked out", a.CurrentCheckout.CheckedOutAt}, [2]string{"Borrower ID", a.CurrentCheckout.CheckedOutByPrincipalID})
	}
	if a.PrimaryPhoto != nil {
		fields = append(fields, [2]string{"Photo", a.PrimaryPhoto.FileName})
	}
	if a.CustomAssetTypeID != nil {
		fields = append(fields, [2]string{"Type ID", *a.CustomAssetTypeID})
	}
	if len(a.Tags) > 0 {
		names := make([]string, 0, len(a.Tags))
		for _, t := range a.Tags {
			names = append(names, t.DisplayName)
		}
		fields = append(fields, [2]string{"Tags", strings.Join(names, ", ")})
	}
	if a.PrintJobID != nil {
		fields = append(fields, [2]string{"Print job", *a.PrintJobID})
	}
	if a.UndoableOperationID != nil {
		fields = append(fields, [2]string{"Undo ID", *a.UndoableOperationID})
	}
	fields = append(fields, [2]string{"Created", a.CreatedAt}, [2]string{"Updated", a.UpdatedAt})
	if err := o.details(fields); err != nil {
		return err
	}
	if len(a.CustomFields) > 0 {
		if _, err := io.WriteString(o.Stdout, "Custom fields:\n"); err != nil {
			return err
		}
		encoder := json.NewEncoder(o.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(a.CustomFields)
	}
	return nil
}
