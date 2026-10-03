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
	case ports.Result[ports.Asset]:
		return o.asset(v.Data)
	default:
		encoder := json.NewEncoder(o.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	}
}
func (o Output) asset(a ports.Asset) error {
	_, err := fmt.Fprintf(o.Stdout, "%s\t%s\t%s\t%s\n", a.ID, a.Kind, strconv.Quote(a.Title), a.Lifecycle)
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
