package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"text/tabwriter"
)

func (o Output) importJobs(jobs []ports.ImportJob) error {
	w := tabwriter.NewWriter(o.Stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "JOB\tSTATUS\tSOURCE\tPHASE\tPROGRESS"); err != nil {
		return err
	}
	for _, v := range jobs {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d/%d\n", strconv.Quote(v.ID), strconv.Quote(v.Status), strconv.Quote(v.Source.Name), strconv.Quote(v.Progress.Phase), v.Progress.Done, v.Progress.Total); err != nil {
			return err
		}
	}
	return w.Flush()
}
func (o Output) importJob(v ports.ImportJob) error {
	fields := [][2]string{{"Job", v.ID}, {"Status", v.Status}, {"Source", v.Source.Name}, {"Source type", v.Source.Type}, {"Created", v.CreatedAt}, {"Updated", v.UpdatedAt}, {"Phase", v.Progress.Phase}, {"Progress", fmt.Sprintf("%d/%d", v.Progress.Done, v.Progress.Total)}, {"Assets created", strconv.FormatInt(v.Counts.AssetsCreated, 10)}, {"Attachments created", strconv.FormatInt(v.Counts.AttachmentsCreated, 10)}, {"Errors", strconv.FormatInt(v.Counts.Errors, 10)}, {"Warnings", strconv.FormatInt(v.Counts.Warnings, 10)}}
	if v.ActorID != nil {
		fields = append(fields, [2]string{"Actor ID", *v.ActorID})
	}
	if v.Actor != nil && v.Actor.Email != nil {
		fields = append(fields, [2]string{"Actor", *v.Actor.Email})
	}
	if v.StartedAt != nil {
		fields = append(fields, [2]string{"Started", *v.StartedAt})
	}
	if v.CompletedAt != nil {
		fields = append(fields, [2]string{"Completed", *v.CompletedAt})
	}
	if v.CancellationMode != nil {
		fields = append(fields, [2]string{"Cancellation mode", *v.CancellationMode})
	}
	if v.Progress.Message != nil {
		fields = append(fields, [2]string{"Progress message", *v.Progress.Message})
	}
	if err := o.details(fields); err != nil {
		return err
	}
	for _, m := range v.Messages {
		f := [][2]string{{"Severity", m.Severity}, {"Code", m.Code}, {"Message", m.Summary}}
		if m.Detail != nil {
			f = append(f, [2]string{"Detail", *m.Detail})
		}
		if err := o.details(f); err != nil {
			return err
		}
	}
	return o.details([][2]string{{"Preview assets", strconv.Itoa(len(v.Preview.Assets))}, {"Preview assets truncated", strconv.FormatBool(v.Preview.AssetsTruncated)}, {"Preview locations", strconv.Itoa(len(v.Preview.Locations))}, {"Preview locations truncated", strconv.FormatBool(v.Preview.LocationsTruncated)}, {"Preview attachments", strconv.Itoa(len(v.Preview.Attachments))}, {"Preview attachments truncated", strconv.FormatBool(v.Preview.AttachmentsTruncated)}, {"Preview fields", strconv.Itoa(len(v.Preview.Fields))}, {"Preview fields truncated", strconv.FormatBool(v.Preview.FieldsTruncated)}, {"Preview tags", strconv.Itoa(len(v.Preview.Tags))}, {"Preview tags truncated", strconv.FormatBool(v.Preview.TagsTruncated)}, {"Preview messages", strconv.Itoa(len(v.Preview.Messages))}, {"Preview messages truncated", strconv.FormatBool(v.Preview.MessagesTruncated)}, {"Full job details", "Use --json for all counts, preview records, progress history and created resources."}})
}
