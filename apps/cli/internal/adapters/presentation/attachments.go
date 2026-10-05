package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (o Output) attachmentDetails(a ports.Attachment) error {
	return o.details([][2]string{{"File", a.FileName}, {"ID", a.ID}, {"Type", a.ContentType}, {"Bytes", strconv.FormatInt(a.SizeBytes, 10)}, {"State", a.Lifecycle}, {"SHA-256", a.SHA256}, {"Created", a.CreatedAt}, {"Asset ID", a.AssetID}, {"Inventory ID", a.InventoryID}, {"Household ID", a.TenantID}})
}
func (o Output) attachmentList(r ports.Result[[]ports.Attachment]) error {
	for _, a := range r.Data {
		if _, err := fmt.Fprintf(o.Stdout, "%s  %s  %s  %d bytes  %s\n", strconv.Quote(a.ID), strconv.Quote(a.FileName), strconv.Quote(a.ContentType), a.SizeBytes, strconv.Quote(a.Lifecycle)); err != nil {
			return err
		}
	}
	return o.pagination(r.Pagination)
}
