package presentation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"strconv"
)

type Output struct {
	Stdout, Stderr io.Writer
	JSON           bool
}

func (o Output) Result(value any) error {
	if o.JSON {
		return json.NewEncoder(o.Stdout).Encode(value)
	}
	switch v := value.(type) {
	case ports.Result[[]ports.Inventory]:
		for _, item := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\n", item.ID, strconv.Quote(item.Name), item.Lifecycle); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[[]ports.Asset]:
		for _, item := range v.Data {
			if err := o.asset(item); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.PrintJobSummary]:
		return o.printJob(v.Data)
	case ports.Result[[]ports.PrintJobSummary]:
		for _, j := range v.Data {
			if err := o.printJob(j); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.RegisteredPrinter]:
		_, err := fmt.Fprintf(o.Stdout, "%s\t%s\tmedia=%s\trevision=%d\n", v.Data.ID, strconv.Quote(v.Data.Name), v.Data.MediaPreset, v.Data.Revision)
		return err
	case ports.Result[[]ports.RegisteredPrinter]:
		for _, p := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", p.ID, strconv.Quote(p.Name), p.Readiness, strconv.Quote(p.MediaName)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[[]ports.LabelTemplate]:
		for _, template := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\tv%d\t%s\n", template.ID, template.Version, strconv.Quote(template.Name)); err != nil {
				return err
			}
		}
		return nil
	case ports.Result[ports.ResolvedLabel]:
		_, err := fmt.Fprintf(o.Stdout, "%s\ttenant=%s\tinventory=%s\t%s\n", v.Data.AssetID, v.Data.TenantID, v.Data.InventoryID, v.Data.Lifecycle)
		return err
	case ports.LabelFileResult:
		_, err := fmt.Fprintf(o.Stdout, "Saved %s (%s), sha256=%s\n", strconv.Quote(v.Path), v.Format, v.SHA256)
		return err
	case ports.Result[ports.Asset]:
		return o.asset(v.Data)
	default:
		encoder := json.NewEncoder(o.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	}
}
func (o Output) printJob(j ports.PrintJobSummary) error {
	_, err := fmt.Fprintf(o.Stdout, "%s\t%s\tprinter=%s\tcopies=%d\n", j.ID, j.Status, j.PrinterID, j.Copies)
	return err
}
func (o Output) asset(a ports.Asset) error {
	_, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", a.ID, a.Kind, strconv.Quote(a.Title), a.Lifecycle)
	if err == nil && a.PrintJobID != "" {
		_, err = fmt.Fprintf(o.Stdout, "Label job: %s\n", a.PrintJobID)
	}
	return err
}
func (o Output) pagination(p *ports.Pagination) error {
	if p != nil && p.HasMore && p.NextCursor != nil {
		_, err := fmt.Fprintf(o.Stdout, "More results: --cursor %s\n", strconv.Quote(*p.NextCursor))
		return err
	}
	return nil
}
func (o Output) Notice(message string) error { _, err := fmt.Fprintln(o.Stderr, message); return err }
func (o Output) Error(category, message string) {
	if o.JSON {
		_ = json.NewEncoder(o.Stderr).Encode(map[string]any{"error": map[string]string{"category": category, "message": message}})
		return
	}
	_, _ = fmt.Fprintf(o.Stderr, "%s: %s\n", category, message)
}

type SilentObserver struct{}

func (SilentObserver) Event(context.Context, string) {}
