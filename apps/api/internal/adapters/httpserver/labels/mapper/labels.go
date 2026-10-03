package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/labels/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

func Label(value printing.LabelView) dto.LabelResponse {
	return dto.LabelResponse{LabelID: string(value.Label.ID), InstanceID: string(value.Label.InstanceID), TenantID: value.Label.TenantID, InventoryID: value.Label.InventoryID, AssetID: value.Label.AssetID, URL: value.URL, LifecycleState: value.LifecycleState}
}
func Media(value dto.Media) printing.MediaSnapshot {
	return printing.MediaSnapshot{PresetID: value.PresetID, Version: value.Version, WidthMicrometers: value.WidthMicrometers, HeightMicrometers: value.HeightMicrometers, MarginsMicrometers: printing.Margins{Left: value.MarginsMicrometers.Left, Right: value.MarginsMicrometers.Right, Top: value.MarginsMicrometers.Top, Bottom: value.MarginsMicrometers.Bottom}, ResolutionDPI: value.ResolutionDPI, RasterWidth: value.RasterWidth, RasterHeight: value.RasterHeight, Orientation: printing.Orientation(value.Orientation), ColorMode: printing.ColorMode(value.ColorMode), CutPolicy: printing.CutPolicy(value.CutPolicy), DisplayRotation: value.DisplayRotation}
}
func Selection(value dto.TemplateSelection) printing.TemplateSelection {
	return printing.TemplateSelection{ID: printing.TemplateID(value.ID), Version: value.Version, Options: printing.TemplateOptions{ShowReference: value.Options.ShowReference}}
}
func Render(value printing.LabelRender) dto.RenderResponse {
	return dto.RenderResponse{ID: string(value.ID), SelectionFingerprint: value.SelectionFingerprint, MediaFingerprint: value.MediaFingerprint, ContentType: value.ContentType, SHA256: value.SHA256, ContentPath: "/tenants/" + value.TenantID + "/inventories/" + value.InventoryID + "/label-renders/" + string(value.ID) + "/content", ExpiresAt: value.ExpiresAt.UTC().Format(time.RFC3339), WidthPixels: value.WidthPixels, HeightPixels: value.HeightPixels, DisplayRotation: value.DisplayRotation}
}
func Template(value printing.TemplateDescriptor) dto.TemplateResponse {
	return dto.TemplateResponse{ID: string(value.ID), Version: value.Version, Name: value.Name, Purpose: value.Purpose, Options: value.Options, Defaults: dto.TemplateOptions{ShowReference: value.Defaults.ShowReference}, Font: value.Font, GlyphCoverage: value.GlyphCoverage, MinimumQRModulePixels: value.MinimumQRModulePixels}
}
