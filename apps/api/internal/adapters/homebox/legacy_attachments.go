package homebox

import (
	"context"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type legacyAttachmentSession struct {
	importer LegacyImporter
	baseURL  string
	token    string
}

func (i LegacyImporter) OpenImportAttachmentSession(ctx context.Context, request ports.ImportSourceRequest) (ports.ImportAttachmentSession, error) {
	if request.SourceType != importplan.SourceLegacyHomebox {
		return nil, errors.New("import source does not provide attachments")
	}
	importer := i.withRequestOptions(request)
	importer.maxAttachmentBytes = normalizedMaxAttachmentBytes(request.MaxAttachmentBytes)
	baseURL, err := normalizeBaseURL(request.BaseURL)
	if err != nil {
		return nil, err
	}
	token, _, err := importer.login(ctx, baseURL, request.Username, request.Password)
	if err != nil {
		return nil, safeLiveSourceError(err)
	}
	return legacyAttachmentSession{importer: importer, baseURL: baseURL, token: token}, nil
}

func (s legacyAttachmentSession) ReadImportAttachment(ctx context.Context, attachment importplan.Attachment) (ports.ImportAttachmentContent, error) {
	itemID := strings.TrimPrefix(strings.TrimSpace(attachment.AssetSourceID), "item:")
	attachmentID := strings.TrimSpace(attachment.SourceID)
	if itemID == "" || attachmentID == "" {
		return ports.ImportAttachmentContent{}, ports.NewImportAttachmentReadError(ports.ImportAttachmentDownloadFailed, errors.New("attachment source identity is invalid"))
	}
	content, contentType, err := s.importer.attachment(ctx, s.baseURL, s.token, itemID, attachmentID)
	if err != nil {
		if errors.Is(err, errLegacyAttachmentTooLarge) {
			return ports.ImportAttachmentContent{}, ports.NewImportAttachmentReadError(ports.ImportAttachmentTooLarge, err)
		}
		return ports.ImportAttachmentContent{}, ports.NewImportAttachmentReadError(ports.ImportAttachmentDownloadFailed, err)
	}
	if !supportedImageType(contentType) {
		return ports.ImportAttachmentContent{}, ports.NewImportAttachmentReadError(ports.ImportAttachmentUnsupportedType, errors.New("attachment content type is unsupported"))
	}
	return ports.ImportAttachmentContent{
		FileName:    safeFileName(attachment.FileName, defaultImageName(contentType)),
		ContentType: contentType,
		Content:     content,
	}, nil
}
