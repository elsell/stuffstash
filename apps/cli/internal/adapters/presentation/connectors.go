package presentation

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"time"
)

func (o Output) printConnector(v ports.PrintConnector) error {
	fields := [][2]string{{"Connector", v.ID}, {"Name", v.Name}, {"State", v.State}, {"Availability", v.Availability}, {"Authorization pending", strconv.FormatBool(v.AuthorizationPending)}, {"Generation", strconv.FormatUint(v.Generation, 10)}}
	if v.LastSeenAt != nil {
		fields = append(fields, [2]string{"Last seen", v.LastSeenAt.Format(time.RFC3339Nano)})
	}
	if v.ReportReceivedAt != nil {
		fields = append(fields, [2]string{"Report received", v.ReportReceivedAt.Format(time.RFC3339Nano)})
	}
	for _, id := range v.PrinterIDs {
		fields = append(fields, [2]string{"Printer", id})
	}
	if v.Report != nil {
		r := v.Report
		fields = append(fields, [2]string{"Platform", r.Platform}, [2]string{"Architecture", r.Architecture}, [2]string{"Version", r.Version}, [2]string{"Commit", r.Commit})
		for _, a := range r.Adapters {
			b, err := json.Marshal(a)
			if err != nil {
				return err
			}
			fields = append(fields, [2]string{"Adapter capability", string(b)})
		}
	}
	return o.details(fields)
}
