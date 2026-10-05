package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"time"
)

func (o Output) printJobRow(j ports.PrintJobSummary) error {
	_, err := fmt.Fprintf(o.Stdout, "%s\t%s\tprinter=%s\tcopies=%d\n", strconv.Quote(j.ID), strconv.Quote(j.Status), strconv.Quote(j.PrinterID), j.Copies)
	return err
}
func (o Output) printJob(j ports.PrintJobSummary) error {
	fields := [][2]string{{"Job", j.ID}, {"Status", j.Status}, {"Kind", j.Kind}, {"Printer", j.PrinterID}, {"Asset", j.AssetID}, {"Previous job", j.Predecessor}, {"Copies", strconv.Itoa(j.Copies)}, {"Revision", strconv.FormatUint(j.Revision, 10)}, {"Requested by", j.RequestedBy}, {"Media fingerprint", j.MediaFingerprint}, {"Created", j.CreatedAt.Format(time.RFC3339Nano)}, {"Updated", j.UpdatedAt.Format(time.RFC3339Nano)}}
	if j.Resolution != nil {
		fields = append(fields, [2]string{"Reported outcome", j.Resolution.ReportedOutcome}, [2]string{"Resolved by", j.Resolution.ResolvedBy}, [2]string{"Resolved", j.Resolution.ResolvedAt.Format(time.RFC3339Nano)})
	}
	if err := o.details(fields); err != nil {
		return err
	}
	for _, a := range j.Attempts {
		fields := [][2]string{{"Attempt", a.ID}, {"Connector", a.ConnectorID}, {"Outcome", a.Outcome}, {"Reason", a.Reason}, {"Completed copies", strconv.Itoa(a.CompletedCopies)}, {"Claimed", a.ClaimedAt.Format(time.RFC3339Nano)}, {"Lease expires", a.LeaseExpiresAt.Format(time.RFC3339Nano)}}
		for _, v := range []struct {
			label string
			when  *time.Time
		}{{"Started", a.StartedAt}, {"Settled", a.SettledAt}, {"Idle confirmed", a.IdleConfirmedAt}} {
			if v.when != nil {
				fields = append(fields, [2]string{v.label, v.when.Format(time.RFC3339Nano)})
			}
		}
		if err := o.details(fields); err != nil {
			return err
		}
	}
	return nil
}
