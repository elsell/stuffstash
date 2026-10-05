package presentation

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"sort"
	"time"
)

func (o Output) auditRecords(r ports.Result[[]ports.AuditRecord]) error {
	for i, v := range r.Data {
		if i > 0 {
			if _, err := io.WriteString(o.Stdout, "\n"); err != nil {
				return err
			}
		}
		fields := [][2]string{{"Record", v.ID}, {"Occurred", v.OccurredAt.Format(time.RFC3339Nano)}, {"Action", v.Action}, {"Source", v.Source}, {"Target type", v.TargetType}, {"Target ID", v.TargetID}, {"Principal ID", v.PrincipalID}, {"Household ID", v.TenantID}}
		if v.InventoryID != nil {
			fields = append(fields, [2]string{"Inventory ID", *v.InventoryID})
		}
		if v.RequestID != nil {
			fields = append(fields, [2]string{"Request ID", *v.RequestID})
		}
		if v.Principal != nil && v.Principal.Email != nil {
			fields = append(fields, [2]string{"Principal", *v.Principal.Email})
		}
		if err := o.details(fields); err != nil {
			return err
		}
		keys := make([]string, 0, len(v.Metadata))
		for key := range v.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := o.details([][2]string{{"Metadata key", key}, {"Value", v.Metadata[key]}}); err != nil {
				return err
			}
		}
	}
	return o.pagination(r.Pagination)
}
