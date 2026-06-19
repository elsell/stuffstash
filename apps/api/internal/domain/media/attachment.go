package media

import (
	"strings"
	"time"
)

type ID string

func NewID(value string) (ID, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return ID(value), true
}

func (id ID) String() string {
	return string(id)
}

type TenantID string

func (id TenantID) String() string {
	return string(id)
}

type InventoryID string

func (id InventoryID) String() string {
	return string(id)
}

type AssetID string

func (id AssetID) String() string {
	return string(id)
}

type StorageKey string

func NewStorageKey(value string) (StorageKey, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "..") {
		return "", false
	}
	return StorageKey(value), true
}

func (key StorageKey) String() string {
	return string(key)
}

type FileName string

func NewFileName(value string) (FileName, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 255 || strings.ContainsAny(value, `/\`) {
		return "", false
	}
	return FileName(value), true
}

func (name FileName) String() string {
	return string(name)
}

type ContentType string

const (
	ContentTypeJPEG ContentType = "image/jpeg"
	ContentTypePNG  ContentType = "image/png"
	ContentTypeWEBP ContentType = "image/webp"
	ContentTypePDF  ContentType = "application/pdf"
)

func NewContentType(value string) (ContentType, bool) {
	switch ContentType(strings.TrimSpace(strings.ToLower(value))) {
	case ContentTypeJPEG:
		return ContentTypeJPEG, true
	case ContentTypePNG:
		return ContentTypePNG, true
	case ContentTypeWEBP:
		return ContentTypeWEBP, true
	case ContentTypePDF:
		return ContentTypePDF, true
	default:
		return "", false
	}
}

func (contentType ContentType) String() string {
	return string(contentType)
}

type SHA256 string

func NewSHA256(value string) (SHA256, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	if len(value) != 64 {
		return "", false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return "", false
		}
	}
	return SHA256(value), true
}

func (hash SHA256) String() string {
	return string(hash)
}

type Attachment struct {
	ID          ID
	TenantID    TenantID
	InventoryID InventoryID
	AssetID     AssetID
	StorageKey  StorageKey
	FileName    FileName
	ContentType ContentType
	SizeBytes   int64
	SHA256      SHA256
	CreatedAt   time.Time
}

func NewAttachment(id ID, tenantID TenantID, inventoryID InventoryID, assetID AssetID, storageKey StorageKey, fileName FileName, contentType ContentType, sizeBytes int64, sha256 SHA256, createdAt time.Time) (Attachment, bool) {
	if sizeBytes <= 0 || createdAt.IsZero() {
		return Attachment{}, false
	}
	return Attachment{
		ID:          id,
		TenantID:    tenantID,
		InventoryID: inventoryID,
		AssetID:     assetID,
		StorageKey:  storageKey,
		FileName:    fileName,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
		SHA256:      sha256,
		CreatedAt:   createdAt.UTC(),
	}, true
}
