package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/labels"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (c *Client) LabelTemplates(ctx context.Context, s ports.Scope) (ports.Result[[]ports.LabelTemplate], error) {
	r, err := read[labelEnvelope[[]ports.LabelTemplate]](c.sdk.GetTenantsByTenantIdInventoriesByInventoryIdLabelTemplates(ctx, s.Tenant, s.Inventory, nil))
	if err != nil {
		return ports.Result[[]ports.LabelTemplate]{}, err
	}
	for i := range r.Data {
		r.Data[i].ShowReference = r.Data[i].Defaults.ShowReference
	}
	return ports.Result[[]ports.LabelTemplate]{Data: r.Data, Schema: r.Schema, Meta: metadata(r.Meta)}, nil
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
	return labelResult(r), nil
}
func (c *Client) LabelMedia(ctx context.Context, s ports.Scope, printer string) ([]printing.Media, error) {
	if printer != "" {
		r, err := c.Printer(ctx, s, printer)
		if err != nil {
			return nil, err
		}
		return []printing.Media{labelProfileMedia(r.Data.Media)}, nil
	}
	r, err := c.PrinterProfiles(ctx, s)
	if err != nil {
		return nil, err
	}
	result := []printing.Media{}
	for _, profile := range r.Data {
		for _, media := range profile.Media {
			result = append(result, labelProfileMedia(media))
		}
	}
	return result, nil
}
func labelProfileMedia(v ports.PrinterMedia) printing.Media {
	return printing.Media{PresetID: v.PresetID, Version: uint32(v.Version), WidthMicrometers: int(v.WidthMicrometers), HeightMicrometers: int(v.HeightMicrometers), Margins: printing.Margins{Left: int(v.MarginsMicrometers.Left), Right: int(v.MarginsMicrometers.Right), Top: int(v.MarginsMicrometers.Top), Bottom: int(v.MarginsMicrometers.Bottom)}, ResolutionDPI: int(v.ResolutionDpi), RasterWidth: int(v.RasterWidth), RasterHeight: int(v.RasterHeight), Orientation: v.Orientation, ColorMode: v.ColorMode, CutPolicy: v.CutPolicy, DisplayRotation: int(v.DisplayRotation)}
}
func (c *Client) RenderLabel(ctx context.Context, s ports.Scope, asset string, selection ports.LabelRenderSelection) (ports.LabelArtifact, error) {
	body := selection.RequestBody
	if len(body) == 0 {
		var err error
		body, err = json.Marshal(labelRenderBody{Media: selection.Media, Format: selection.Format, Template: labelRenderTemplate{ID: selection.TemplateID, Version: selection.TemplateVersion, Options: ports.LabelTemplateDefaults{ShowReference: selection.ShowReference}}})
		if err != nil {
			return ports.LabelArtifact{}, ports.Failure("input", "The render selection could not be encoded.")
		}
	}
	if _, err := read[generated.SuccessEnvelopeLabelResponse](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdLabel(ctx, s.Tenant, s.Inventory, asset, nil)); err != nil {
		return ports.LabelArtifact{}, err
	}
	rendered, err := read[labelEnvelope[ports.LabelRenderMetadata]](c.sdk.PostTenantsByTenantIdInventoriesByInventoryIdAssetsByAssetIdLabelRendersWithBody(ctx, s.Tenant, s.Inventory, asset, nil, "application/json", bytes.NewReader(body)))
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
	response, err := c.sdk.ListTenantsByTenantIdInventoriesByInventoryIdLabelRendersByRenderIdContent(ctx, s.Tenant, s.Inventory, rendered.Data.ID, nil)
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
	if len(content) > maximumBytes || !bytes.HasPrefix(content, signature) || !strings.EqualFold(checksum, rendered.Data.SHA256) {
		return ports.LabelArtifact{}, ports.Failure("protocol", "label content integrity check failed")
	}
	return ports.LabelArtifact{Content: content, Format: selection.Format, SHA256: checksum, Render: ports.Result[ports.LabelRenderMetadata]{Data: rendered.Data, Schema: rendered.Schema, Meta: metadata(rendered.Meta)}}, nil
}

var _ ports.LabelsAPI = (*Client)(nil)
