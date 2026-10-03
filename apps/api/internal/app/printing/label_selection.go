package printing

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	label "github.com/stuffstash/stuff-stash/internal/domain/printing"
)

// ValidateDefaultLabelSelection uses the same renderer as real output, but does
// not provision identity, persist artifacts, or claim physical-print readiness.
func (s *LabelService) ValidateDefaultLabelSelection(ctx context.Context, selection label.TemplateSelection, media *label.MediaSnapshot) error {
	if s == nil || s.deps.Templates == nil || s.deps.Renderer == nil {
		return ErrLabelsUnavailable
	}
	found := false
	for _, descriptor := range s.deps.Templates.Templates() {
		if descriptor.ID == selection.ID && descriptor.Version == selection.Version {
			found = true
			break
		}
	}
	if !found {
		return apperrors.ErrInvalidInput
	}
	if media == nil {
		return nil
	}
	const sampleID = "00000000000000000000000000"
	content := label.ContentSnapshot{QRURL: strings.TrimRight(s.deps.BaseURL, "/") + "/l/v1/" + sampleID + "/" + sampleID, Title: "Sample item", Reference: sampleID}
	_, err := s.deps.Renderer.Render(ctx, label.RenderRequest{Content: content, Template: selection, Media: *media, Format: label.FormatPNG})
	if err != nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}
