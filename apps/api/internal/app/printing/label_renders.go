package printing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	label "github.com/stuffstash/stuff-stash/internal/domain/printing"
)

type LabelRenderInput struct {
	Template label.TemplateSelection
	Media    label.MediaSnapshot
	Format   label.Format
}

func (s *LabelService) Render(ctx context.Context, scope LabelScope, assetID asset.ID, input LabelRenderInput) (label.LabelRender, error) {
	view, err := s.Get(ctx, scope, assetID)
	if err != nil {
		return label.LabelRender{}, err
	}
	item, err := s.asset(ctx, scope, assetID)
	if err != nil {
		return label.LabelRender{}, err
	}
	if s.deps.Renderer == nil || s.deps.Renders == nil || s.deps.RenderTTL <= 0 || s.deps.MaxRenderBytes <= 0 {
		return label.LabelRender{}, ErrLabelsUnavailable
	}
	request := label.RenderRequest{Content: label.ContentSnapshot{QRURL: view.URL, Title: item.Title.String(), Reference: string(view.Label.ID)}, Template: input.Template, Media: input.Media, Format: input.Format}
	rendered, err := s.deps.Renderer.Render(ctx, request)
	if err != nil {
		return label.LabelRender{}, fmt.Errorf("%w: label cannot be rendered", apperrors.ErrInvalidInput)
	}
	if len(rendered.Content) > s.deps.MaxRenderBytes {
		return label.LabelRender{}, apperrors.ErrInvalidInput
	}
	// Recheck after rendering: a user whose permission changed during CPU work
	// must not receive a newly created artifact. Content reads check again too.
	if _, err := s.asset(ctx, scope, assetID); err != nil {
		return label.LabelRender{}, err
	}
	selected, _ := json.Marshal(request)
	fingerprint := sha256.Sum256(selected)
	now := s.deps.Clock.Now()
	record := label.LabelRender{ID: label.RenderID(s.deps.IDs.NewID()), TenantID: scope.TenantID.String(), InventoryID: scope.InventoryID.String(), AssetID: assetID.String(), LabelID: view.Label.ID, SelectionFingerprint: hex.EncodeToString(fingerprint[:]), MediaFingerprint: input.Media.Fingerprint(), ContentType: rendered.ContentType, SHA256: rendered.SHA256, Content: rendered.Content, WidthPixels: rendered.WidthPixels, HeightPixels: rendered.HeightPixels, DisplayRotation: rendered.DisplayRotation, CreatedAt: now, ExpiresAt: now.Add(s.deps.RenderTTL)}
	auditRecord, err := appsupport.NewAuditRecord(s.deps.IDs, s.deps.Clock, s.auditInput(scope, audit.ActionLabelRendered, assetID.String()))
	if err != nil {
		return label.LabelRender{}, err
	}
	if err := s.deps.Renders.SaveLabelRender(ctx, record, auditRecord); err != nil {
		return label.LabelRender{}, err
	}
	return record, nil
}
func (s *LabelService) Content(ctx context.Context, scope LabelScope, id label.RenderID) (label.LabelRender, error) {
	if err := s.access(ctx, scope); err != nil {
		return label.LabelRender{}, err
	}
	if s.deps.Renders == nil {
		return label.LabelRender{}, ErrLabelsUnavailable
	}
	record, found, err := s.deps.Renders.LabelRenderByID(ctx, scope.TenantID, scope.InventoryID, id)
	if err != nil {
		return label.LabelRender{}, err
	}
	if !found || !s.deps.Clock.Now().Before(record.ExpiresAt) {
		return label.LabelRender{}, apperrors.ErrNotFound
	}
	if _, err := s.asset(ctx, scope, asset.ID(record.AssetID)); err != nil {
		return label.LabelRender{}, err
	}
	if err := s.readAudit(ctx, scope, audit.ActionLabelContentDownloaded, record.AssetID); err != nil {
		return label.LabelRender{}, err
	}
	return record, nil
}
func (s *LabelService) Templates(ctx context.Context, scope LabelScope) ([]label.TemplateDescriptor, error) {
	if err := s.access(ctx, scope); err != nil {
		return nil, err
	}
	if s.deps.Templates == nil {
		return nil, ErrLabelsUnavailable
	}
	if err := s.readAudit(ctx, scope, audit.ActionLabelTemplatesListed, ""); err != nil {
		return nil, err
	}
	return s.deps.Templates.Templates(), nil
}
func (s *LabelService) PurgeExpired(ctx context.Context) error {
	if s == nil || s.deps.Renders == nil {
		return nil
	}
	return s.deps.Renders.PurgeExpiredLabelRenders(ctx, s.deps.Clock.Now())
}
