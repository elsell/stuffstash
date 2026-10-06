package presentation

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"sort"
	"time"
)

func (o Output) activity(r ports.Result[[]ports.Activity]) error {
	for i, v := range r.Data {
		if i > 0 {
			if _, err := io.WriteString(o.Stdout, "\n"); err != nil {
				return err
			}
		}
		fields := [][2]string{{"Occurred", v.OccurredAt.Format(time.RFC3339Nano)}, {"Action", v.Action}, {"Category", v.Category}, {"Event", v.ID}, {"Principal ID", v.PrincipalID}, {"Source", v.Source}}
		if v.Principal != nil && v.Principal.Email != nil {
			fields = append(fields, [2]string{"Principal", *v.Principal.Email})
		}
		if v.RequestID != nil {
			fields = append(fields, [2]string{"Request ID", *v.RequestID})
		}
		if v.Undo != nil {
			fields = append(fields, [2]string{"Undo operation", v.Undo.OperationID}, [2]string{"Undo status", v.Undo.Status})
		}
		for _, c := range v.Changes {
			fields = append(fields, [2]string{"Changed field", c.Field})
			if c.PreviousValue != nil {
				fields = append(fields, [2]string{"Previous value", *c.PreviousValue})
			}
			if c.CurrentValue != nil {
				fields = append(fields, [2]string{"Current value", *c.CurrentValue})
			}
		}
		keys := make([]string, 0, len(v.TechnicalMetadata))
		for k := range v.TechnicalMetadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fields = append(fields, [2]string{"Metadata key", k}, [2]string{"Value", v.TechnicalMetadata[k]})
		}
		if err := o.details(fields); err != nil {
			return err
		}
	}
	return o.pagination(r.Pagination)
}
