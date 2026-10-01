package media

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"golang.org/x/image/webp"
)

func encodeAttachmentCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, id media.ID) *string {
	return appsupport.EncodePageCursor("attachments", tenantID.String()+":"+inventoryID.String()+":"+assetID.String(), id.String())
}

func decodeAttachmentCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID, cursor string) (media.ID, error) {
	decoded, err := appsupport.DecodePageCursor("attachments", tenantID.String()+":"+inventoryID.String()+":"+assetID.String(), cursor)
	if err != nil {
		return "", err
	}
	if decoded == "" {
		return "", nil
	}
	id, ok := media.NewID(decoded)
	if !ok {
		return "", apperrors.ErrInvalidInput
	}
	return id, nil
}

func (a AttachmentService) ensureActiveAssetForAttachment(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) error {
	item, found, err := a.deps.Assets.AssetByID(ctx, tenantID, inventoryID, assetID)
	if err != nil {
		return err
	}
	if !found || item.LifecycleState != asset.LifecycleStateActive {
		return apperrors.ErrNotFound
	}
	return nil
}

func (a AttachmentService) ensureReadableAssetForAttachment(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) error {
	item, found, err := a.deps.Assets.AssetByID(ctx, tenantID, inventoryID, assetID)
	if err != nil {
		return err
	}
	if !found {
		return apperrors.ErrNotFound
	}
	switch item.LifecycleState {
	case asset.LifecycleStateActive, asset.LifecycleStateArchived:
		return nil
	default:
		return apperrors.ErrNotFound
	}
}

func contentMatchesType(contentType media.ContentType, content []byte) bool {
	switch contentType {
	case media.ContentTypePNG:
		return len(content) >= 8 &&
			content[0] == 0x89 &&
			content[1] == 'P' &&
			content[2] == 'N' &&
			content[3] == 'G' &&
			content[4] == '\r' &&
			content[5] == '\n' &&
			content[6] == 0x1a &&
			content[7] == '\n' &&
			imageContentDecodes(content, png.Decode)
	case media.ContentTypeJPEG:
		return len(content) >= 3 &&
			content[0] == 0xff &&
			content[1] == 0xd8 &&
			content[2] == 0xff &&
			imageContentDecodes(content, jpeg.Decode)
	case media.ContentTypeWEBP:
		return len(content) >= 12 &&
			string(content[0:4]) == "RIFF" &&
			string(content[8:12]) == "WEBP" &&
			imageContentDecodes(content, webp.Decode)
	case media.ContentTypePDF:
		return len(content) >= 5 && string(content[0:5]) == "%PDF-"
	default:
		return false
	}
}

func imageContentDecodes(content []byte, decode func(io.Reader) (image.Image, error)) bool {
	decoded, err := decode(bytes.NewReader(content))
	if err != nil {
		return false
	}
	bounds := decoded.Bounds()
	return bounds.Dx() > 0 && bounds.Dy() > 0
}
