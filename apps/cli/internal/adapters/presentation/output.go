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
	case ports.Result[[]ports.Tag]:
		for _, item := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", strconv.Quote(item.ID), strconv.Quote(item.DisplayName), strconv.Quote(item.Key), strconv.Quote(item.Lifecycle)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[[]ports.CheckedOutAsset]:
		for _, item := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s  %s  %s  %s\n", strconv.Quote(item.Asset.ID), strconv.Quote(item.Asset.Title), strconv.Quote(item.Checkout.CheckedOutByPrincipalID), strconv.Quote(item.Checkout.CheckedOutAt)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.Checkout]:
		return o.checkoutDetails(v.Data)
	case ports.Result[[]ports.Checkout]:
		for i, item := range v.Data {
			if i > 0 {
				if _, err := io.WriteString(o.Stdout, "\n"); err != nil {
					return err
				}
			}
			if err := o.checkoutDetails(item); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
	case ports.Result[ports.Tag]:
		fields := [][2]string{{"Tag", v.Data.DisplayName}, {"ID", v.Data.ID}, {"Key", v.Data.Key}, {"State", v.Data.Lifecycle}}
		if v.Data.Color != nil {
			fields = append(fields, [2]string{"Color", *v.Data.Color})
		}
		return o.details(fields)
	case ports.Result[ports.Principal]:
		fields := [][2]string{{"ID", v.Data.ID}}
		if v.Data.DisplayName != nil {
			fields = append(fields, [2]string{"Name", *v.Data.DisplayName})
		}
		if v.Data.Email != nil {
			fields = append(fields, [2]string{"Email", *v.Data.Email})
		}
		return o.details(fields)
	case ports.Result[ports.Tenant]:
		return o.details([][2]string{{"Household", v.Data.Name}, {"ID", v.Data.ID}, {"State", v.Data.Lifecycle}, {"Access", v.Data.Access.Relationship}})
	case ports.Result[ports.Inventory]:
		return o.details([][2]string{{"Inventory", v.Data.Name}, {"ID", v.Data.ID}, {"Household ID", v.Data.TenantID}, {"State", v.Data.Lifecycle}, {"Access", v.Data.Access.Relationship}})
	case ports.Result[[]ports.Tenant]:
		for _, item := range v.Data {
			if _, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\n", strconv.Quote(item.ID), strconv.Quote(item.Name), strconv.Quote(item.Lifecycle)); err != nil {
				return err
			}
		}
		return o.pagination(v.Pagination)
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
		return o.assetDetails(v.Data)
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
	_, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", strconv.Quote(a.ID), strconv.Quote(a.Kind), strconv.Quote(a.Title), strconv.Quote(a.Lifecycle))
	if err == nil && a.PrintJobID != nil {
		_, err = fmt.Fprintf(o.Stdout, "Label job: %s\n", strconv.Quote(*a.PrintJobID))
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

func (o Output) details(fields [][2]string) error {
	for _, field := range fields {
		if _, err := fmt.Fprintf(o.Stdout, "%-13s %s\n", field[0]+":", strconv.Quote(field[1])); err != nil {
			return err
		}
	}
	return nil
}
