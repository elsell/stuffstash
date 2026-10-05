package httpapi

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) PrinterProfiles(ctx context.Context, s ports.Scope) (ports.Result[[]ports.PrinterProfile], error) {
	r, err := read[generated.SuccessEnvelopeListPrinterProfile](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrinterProfiles(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return ports.Result[[]ports.PrinterProfile]{}, err
	}
	var profiles []ports.PrinterProfile
	if r.Data.GetOrEmpty() != nil {
		profiles = make([]ports.PrinterProfile, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		var media []ports.PrinterMedia
		if v.Media.GetOrEmpty() != nil {
			media = make([]ports.PrinterMedia, 0, len(v.Media.GetOrEmpty()))
		}
		for _, m := range v.Media.GetOrEmpty() {
			media = append(media, printerMedia(m))
		}
		profiles = append(profiles, ports.PrinterProfile{AdapterID: v.AdapterId, Name: v.Name, Transport: v.Transport, PhysicallyVerified: v.PhysicallyVerified, SupportedPlatforms: v.SupportedPlatforms.GetOrEmpty(), Media: media})
	}
	return ports.Result[[]ports.PrinterProfile]{Data: profiles, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
}
