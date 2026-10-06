package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"text/tabwriter"
	"time"
)

func (o Output) workflows(v ports.Result[[]ports.WorkflowHead]) error {
	if len(v.Data) == 0 {
		if _, err := fmt.Fprintln(o.Stdout, "No workflows found."); err != nil {
			return err
		}
		return o.pagination(v.Pagination)
	}
	w := tabwriter.NewWriter(o.Stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "ID\tNAME\tLATEST REVISION\tACTIVE REVISION"); err != nil {
		return err
	}
	for _, v := range v.Data {
		active := "Not selected"
		if v.ActiveRevisionId != nil {
			active = *v.ActiveRevisionId
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%d (%s)\t%s\n", strconv.Quote(v.Id), strconv.Quote(v.Name), v.LatestRevision, strconv.Quote(v.LatestRevisionId), strconv.Quote(active)); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return o.pagination(v.Pagination)
}
func (o Output) workflowRevisions(v ports.Result[[]ports.WorkflowRevision]) error {
	if len(v.Data) == 0 {
		if _, err := fmt.Fprintln(o.Stdout, "No workflow revisions found."); err != nil {
			return err
		}
		return o.pagination(v.Pagination)
	}
	w := tabwriter.NewWriter(o.Stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "ID\tREVISION\tNAME\tAUTHOR\tCREATED"); err != nil {
		return err
	}
	for _, v := range v.Data {
		if _, err := fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n", strconv.Quote(v.Id), v.Number, strconv.Quote(v.Definition.Name), strconv.Quote(v.AuthorId), v.CreatedAt.Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return o.pagination(v.Pagination)
}
func (o Output) workflowRevision(v ports.WorkflowRevision) error {
	fields := [][2]string{{"Workflow", v.WorkflowId}, {"Revision ID", v.Id}, {"Revision", strconv.FormatInt(v.Number, 10)}, {"Name", v.Definition.Name}, {"Author", v.AuthorId}, {"Created", v.CreatedAt.Format(time.RFC3339Nano)}}
	for _, f := range []struct {
		label string
		value *string
	}{{"Instructions", v.Definition.Instructions}, {"Provider profile", v.Definition.ProviderProfileId}, {"Settings migration", v.SettingsMigration}} {
		if f.value != nil {
			fields = append(fields, [2]string{f.label, *f.value})
		}
	}
	b := v.Definition.Budget
	fields = append(fields, [2]string{"Elapsed seconds", strconv.FormatInt(b.ElapsedSeconds, 10)}, [2]string{"Follow-up turns", strconv.FormatInt(b.FollowUpTurns, 10)}, [2]string{"Model calls", strconv.FormatInt(b.ModelCalls, 10)}, [2]string{"Tool calls", strconv.FormatInt(b.ToolCalls, 10)})
	return o.details(fields)
}
func (o Output) workflowSelection(v *ports.WorkflowSelection) error {
	if v == nil {
		_, err := fmt.Fprintln(o.Stdout, "No workflow selected.")
		return err
	}
	return o.details([][2]string{{"Workflow", v.WorkflowId}, {"Revision", v.RevisionId}})
}
