package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

func (r Runner) choosePairingBindings(ctx context.Context, o Options, token string, review ports.PairingReview) ([]ports.PairingBinding, error) {
	if r.PrintingAPI == nil {
		return nil, ports.Failure("configuration", "Printer selection is not available. Use structured input with explicit bindings.")
	}
	api, err := r.PrintingAPI(o.Server, token)
	if err != nil {
		return nil, err
	}
	var printers []ports.RegisteredPrinter
	p := ports.Page{Limit: 100}
	seen := map[string]bool{}
	for {
		result, err := api.RegisteredPrinters(ctx, o.Scope, p)
		if err != nil {
			return nil, err
		}
		printers = append(printers, result.Data...)
		if result.Pagination == nil || !result.Pagination.HasMore {
			break
		}
		cursor := result.Pagination.NextCursor
		if cursor == nil || *cursor == "" || seen[*cursor] {
			return nil, ports.Failure("protocol", "Printer pagination did not advance. List printers and try again.")
		}
		seen[*cursor] = true
		p.Cursor = *cursor
	}
	var bindings []ports.PairingBinding
	used := map[string]bool{}
	for _, candidate := range review.Candidates {
		choices := []ports.Choice{{ID: "skip", Label: "Skip this candidate"}}
		for _, printer := range printers {
			if !printer.Retired && printer.AdapterID == candidate.AdapterID && !used[printer.ID] {
				choices = append(choices, ports.Choice{ID: "printer:" + printer.ID, Label: strconv.Quote(printer.Name), Detail: "Printer " + strconv.Quote(printer.ID) + ". Adapter " + strconv.Quote(printer.AdapterID)})
			}
		}
		selected, err := r.Picker.Pick(ctx, "Bind candidate "+strconv.Quote(candidate.ID)+" ("+strconv.Quote(candidate.Name)+")", choices)
		if err != nil {
			return nil, err
		}
		if selected == "skip" {
			continue
		}
		valid := false
		for _, c := range choices {
			if c.ID == selected {
				valid = true
				break
			}
		}
		if !valid {
			return nil, ports.Failure("usage", "Select one of the displayed printer bindings.")
		}
		id := strings.TrimPrefix(selected, "printer:")
		bindings = append(bindings, ports.PairingBinding{CandidateID: candidate.ID, PrinterID: id})
		used[id] = true
	}
	return bindings, validatePairingBindings(bindings)
}
