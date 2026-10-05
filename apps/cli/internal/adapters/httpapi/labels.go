package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/labels"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) LabelTemplates(ctx context.Context, s ports.Scope) (ports.Result[[]ports.LabelTemplate], error) {
	r, err := read[generated.SuccessEnvelopeListTemplateResponse](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdLabelTemplates(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return ports.Result[[]ports.LabelTemplate]{}, err
	}
	result := ports.Result[[]ports.LabelTemplate]{Schema: r.Schema, Meta: metadata(r.Meta)}
	if r.Data.GetOrEmpty() != nil {
		result.Data = make([]ports.LabelTemplate, 0, len(r.Data.GetOrEmpty()))
	}
	for _, v := range r.Data.GetOrEmpty() {
		result.Data = append(result.Data, ports.LabelTemplate{Defaults: ports.LabelTemplateDefaults{ShowReference: v.Defaults.ShowReference}, Font: v.Font, GlyphCoverage: v.GlyphCoverage, MinimumQRModulePixels: v.MinimumQRModulePixels, Options: v.Options.GetOrEmpty(), ID: v.Id, Version: uint32(v.Version), Name: v.Name, Purpose: v.Purpose, ShowReference: v.Defaults.ShowReference})
	}
	return result, nil
}
func (c *Client) ResolveLabel(ctx context.Context, ref labels.Reference) (ports.Result[ports.ResolvedLabel], error) {
	instance, err := read[generated.SuccessEnvelopeInstanceResponse](c.sdk.GetInstance(ctx))
	if err != nil {
		return ports.Result[ports.ResolvedLabel]{}, err
	}
	if instance.Data.ProtocolVersion != 1 || instance.Data.InstanceId != ref.Instance {
		return ports.Result[ports.ResolvedLabel]{}, ports.Failure("not_found", "label does not belong to this instance")
	}
	r, err := read[generated.SuccessEnvelopeLabelResponse](c.sdk.GetLabelsV1ByInstanceIdByLabelId(ctx, ref.Instance, ref.Label, nil))
	if err != nil {
		return ports.Result[ports.ResolvedLabel]{}, err
	}
	if r.Data.InstanceId != ref.Instance || r.Data.LabelId != ref.Label {
		return ports.Result[ports.ResolvedLabel]{}, ports.Failure("protocol", "label identity did not match")
	}
	return ports.Result[ports.ResolvedLabel]{Data: ports.ResolvedLabel{TenantID: r.Data.TenantId, InventoryID: r.Data.InventoryId, AssetID: r.Data.AssetId, Lifecycle: r.Data.LifecycleState}}, nil
}
func (c *Client) LabelMedia(ctx context.Context, s ports.Scope, printer string) ([]printing.Media, error) {
	if printer != "" {
		r, err := read[generated.SuccessEnvelopePrinter](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrintersByPrinterId(ctx, s.Tenant, s.Inventory, printer, nil))
		if err != nil {
			return nil, err
		}
		return []printing.Media{labelProfileMedia(r.Data.Media)}, nil
	}
	r, err := read[generated.SuccessEnvelopeListPrinterProfile](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdPrinterProfiles(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return nil, err
	}
	result := []printing.Media{}
	for _, profile := range r.Data.GetOrEmpty() {
		for _, media := range profile.Media.GetOrEmpty() {
			result = append(result, labelProfileMedia(media))
		}
	}
	return result, nil
}
func labelProfileMedia(v generated.MediaProfile) printing.Media {
	return printing.Media{PresetID: v.PresetId, Version: uint32(v.Version), WidthMicrometers: int(v.WidthMicrometers), HeightMicrometers: int(v.HeightMicrometers), Margins: printing.Margins{Left: int(v.MarginsMicrometers.Left), Right: int(v.MarginsMicrometers.Right), Top: int(v.MarginsMicrometers.Top), Bottom: int(v.MarginsMicrometers.Bottom)}, ResolutionDPI: int(v.ResolutionDpi), RasterWidth: int(v.RasterWidth), RasterHeight: int(v.RasterHeight), Orientation: v.Orientation, ColorMode: v.ColorMode, CutPolicy: v.CutPolicy, DisplayRotation: int(v.DisplayRotation)}
}
func labelMedia(v printing.Media) generated.Media {
	version := int32(v.Version)
	return generated.Media{PresetId: &v.PresetID, Version: &version, WidthMicrometers: int64(v.WidthMicrometers), HeightMicrometers: int64(v.HeightMicrometers), MarginsMicrometers: generated.Margins{Left: int64(v.Margins.Left), Right: int64(v.Margins.Right), Top: int64(v.Margins.Top), Bottom: int64(v.Margins.Bottom)}, ResolutionDpi: int64(v.ResolutionDPI), RasterWidth: int64(v.RasterWidth), RasterHeight: int64(v.RasterHeight), Orientation: v.Orientation, ColorMode: v.ColorMode, CutPolicy: v.CutPolicy, DisplayRotation: int64(v.DisplayRotation)}
}
func (c *Client) RenderLabel(ctx context.Context, s ports.Scope, asset string, selection ports.LabelRenderSelection) (ports.LabelArtifact, error) {
	if _, err := read[generated.SuccessEnvelopeLabelResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdLabel(ctx, s.Tenant, s.Inventory, asset, nil)); err != nil {
		return ports.LabelArtifact{}, err
	}
	rendered, err := read[generated.SuccessEnvelopeRenderResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdLabelRenders(ctx, s.Tenant, s.Inventory, asset, nil, generated.RenderInputBody{Format: generated.RenderInputBodyFormat(selection.Format), Media: labelMedia(selection.Media), Template: generated.TemplateSelection{Id: selection.TemplateID, Version: int32(selection.TemplateVersion), Options: generated.TemplateOptions{ShowReference: selection.ShowReference}}}))
	if err != nil {
		return ports.LabelArtifact{}, err
	}
	expectedType := "image/png"
	signature := []byte("\x89PNG\r\n\x1a\n")
	if selection.Format == "pdf" {
		expectedType = "application/pdf"
		signature = []byte("%PDF-")
	}
	if rendered.Data.ContentType != expectedType {
		return ports.LabelArtifact{}, ports.Failure("protocol", "unexpected label format")
	}
	response, err := c.sdk.ListTenantsByTenantIdInventoriesByInventoryIdLabelRendersByRenderIdContent(ctx, s.Tenant, s.Inventory, rendered.Data.Id, nil)
	if err != nil {
		return ports.LabelArtifact{}, ports.Failure("network", "label download failed")
	}
	if response.StatusCode != 200 {
		_, err = read[struct{}](response, nil)
		if err == nil {
			err = ports.Failure("protocol", "unexpected label response")
		}
		return ports.LabelArtifact{}, err
	}
	defer response.Body.Close()
	contentType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || contentType != expectedType {
		return ports.LabelArtifact{}, ports.Failure("protocol", "unexpected label content type")
	}
	const maximumBytes = 16 << 20
	content, err := io.ReadAll(io.LimitReader(response.Body, maximumBytes+1))
	if err != nil {
		return ports.LabelArtifact{}, ports.Failure("network", "label download interrupted")
	}
	digest := sha256.Sum256(content)
	checksum := hex.EncodeToString(digest[:])
	if len(content) > maximumBytes || !bytes.HasPrefix(content, signature) || !strings.EqualFold(checksum, rendered.Data.Sha256) {
		return ports.LabelArtifact{}, ports.Failure("protocol", "label content integrity check failed")
	}
	return ports.LabelArtifact{Content: content, Format: selection.Format, SHA256: checksum}, nil
}

var _ ports.LabelsAPI = (*Client)(nil)
