package httpapi

import (
	"context"
	"math"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) PrinterMediaPresets(ctx context.Context, s ports.Scope, adapterID string) ([]ports.PrinterMediaPreset, error) {
	response, err := read[generated.SuccessEnvelopeListPrinterProfile](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrinterProfiles(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return nil, err
	}
	presets := []ports.PrinterMediaPreset{}
	for _, profile := range response.Data.GetOrEmpty() {
		if profile.AdapterId != adapterID {
			continue
		}
		for _, media := range profile.Media.GetOrEmpty() {
			if media.Version <= 0 {
				return nil, ports.Failure("protocol", "printer catalog contains an invalid media version")
			}
			presets = append(presets, ports.PrinterMediaPreset{ID: media.PresetId, Version: uint32(media.Version)})
		}
	}
	return presets, nil
}
func (c *Client) ConfigurePrinterMedia(ctx context.Context, s ports.Scope, id string, revision uint64, preset ports.PrinterMediaPreset) (ports.Result[ports.RegisteredPrinter], error) {
	if revision == 0 || revision > math.MaxInt64 || preset.Version == 0 || preset.Version > math.MaxInt32 {
		return ports.Result[ports.RegisteredPrinter]{}, ports.Failure("protocol", "printer configuration has an invalid revision or media version")
	}
	version := int32(preset.Version)
	response, err := read[generated.SuccessEnvelopePrinter](c.sdk.PatchTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterId(ctx, s.Tenant, s.Inventory, id, nil, generated.UpdatePrinterBody{Revision: int64(revision), PresetId: &preset.ID, PresetVersion: &version}))
	if err != nil {
		return ports.Result[ports.RegisteredPrinter]{}, err
	}
	return ports.Result[ports.RegisteredPrinter]{Data: humanPrinter(response.Data), Schema: response.Schema, Meta: metadata(response.Meta)}, nil
}
