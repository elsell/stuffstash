package blobstore

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type S3DirectAttachmentUploader struct {
	store S3Store
}

func NewS3DirectAttachmentUploader(store S3Store) *S3DirectAttachmentUploader {
	return &S3DirectAttachmentUploader{store: store}
}

func (u *S3DirectAttachmentUploader) CreateDirectAttachmentUpload(ctx context.Context, request ports.DirectAttachmentUploadRequest) (ports.DirectAttachmentUpload, error) {
	if u == nil || u.store.presignClient == nil || len(u.store.directUploadSigningKey) == 0 || strings.TrimSpace(request.UploadID) == "" || request.ExpiresAt.IsZero() || !request.ExpiresAt.After(time.Now().UTC()) {
		return ports.DirectAttachmentUpload{}, ports.ErrDirectUploadInvalid
	}

	policy := minio.NewPostPolicy()
	for _, setPolicy := range []func() error{
		func() error { return policy.SetBucket(u.store.bucket) },
		func() error { return policy.SetKey(request.StorageKey.String()) },
		func() error { return policy.SetExpires(request.ExpiresAt) },
		func() error { return policy.SetContentType(request.ContentType.String()) },
		func() error { return policy.SetContentLengthRange(request.SizeBytes, request.SizeBytes) },
	} {
		if err := setPolicy(); err != nil {
			return ports.DirectAttachmentUpload{}, fmt.Errorf("%w: %v", ports.ErrDirectUploadInvalid, err)
		}
	}

	uploadURL, formFields, err := u.store.presignClient.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return ports.DirectAttachmentUpload{}, err
	}
	uploadID, err := encodeDirectUploadToken(request, u.store.directUploadSigningKey)
	if err != nil {
		return ports.DirectAttachmentUpload{}, err
	}

	return ports.DirectAttachmentUpload{
		UploadID:     uploadID,
		AttachmentID: request.AttachmentID,
		Method:       "POST",
		URL:          uploadURL.String(),
		Headers:      map[string]string{},
		FormFields:   copyStringMap(formFields),
		ExpiresAt:    request.ExpiresAt,
	}, nil
}

func (u *S3DirectAttachmentUploader) CompleteDirectAttachmentUpload(ctx context.Context, uploadID string) (ports.CompletedDirectAttachmentUpload, error) {
	if u == nil || u.store.client == nil || len(u.store.directUploadSigningKey) == 0 {
		return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadInvalid
	}

	request, err := decodeDirectUploadToken(uploadID, u.store.directUploadSigningKey)
	if err != nil {
		return ports.CompletedDirectAttachmentUpload{}, err
	}
	request.UploadID = uploadID
	if !request.ExpiresAt.After(time.Now().UTC()) {
		return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadExpired
	}

	stat, err := u.store.client.StatObject(ctx, u.store.bucket, request.StorageKey.String(), minio.StatObjectOptions{})
	if err != nil {
		if errors.Is(mapS3Error(err), ports.ErrBlobNotFound) {
			return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadIncomplete
		}
		return ports.CompletedDirectAttachmentUpload{}, err
	}
	if stat.Size != request.SizeBytes || strings.TrimSpace(stat.ContentType) != request.ContentType.String() {
		return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadMismatch
	}

	data, err := u.store.GetBlob(ctx, request.StorageKey)
	if err != nil {
		if errors.Is(err, ports.ErrBlobNotFound) {
			return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadIncomplete
		}
		return ports.CompletedDirectAttachmentUpload{}, err
	}
	if int64(len(data)) != request.SizeBytes {
		return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadMismatch
	}
	hashBytes := sha256.Sum256(data)
	hash, ok := media.NewSHA256(hex.EncodeToString(hashBytes[:]))
	if !ok {
		return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadMismatch
	}

	return ports.CompletedDirectAttachmentUpload{
		UploadID:     request.UploadID,
		AttachmentID: request.AttachmentID,
		TenantID:     request.TenantID,
		InventoryID:  request.InventoryID,
		AssetID:      request.AssetID,
		StorageKey:   request.StorageKey,
		FileName:     request.FileName,
		ContentType:  request.ContentType,
		SizeBytes:    request.SizeBytes,
		SHA256:       hash,
		ExpiresAt:    request.ExpiresAt,
	}, nil
}

type directUploadTokenPayload struct {
	UploadID     string `json:"uploadId"`
	AttachmentID string `json:"attachmentId"`
	TenantID     string `json:"tenantId"`
	InventoryID  string `json:"inventoryId"`
	AssetID      string `json:"assetId"`
	StorageKey   string `json:"storageKey"`
	FileName     string `json:"fileName"`
	ContentType  string `json:"contentType"`
	SizeBytes    int64  `json:"sizeBytes"`
	ExpiresAt    string `json:"expiresAt"`
}

func encodeDirectUploadToken(request ports.DirectAttachmentUploadRequest, signingKey []byte) (string, error) {
	payload := directUploadTokenPayload{
		UploadID:     request.UploadID,
		AttachmentID: request.AttachmentID.String(),
		TenantID:     request.TenantID.String(),
		InventoryID:  request.InventoryID.String(),
		AssetID:      request.AssetID.String(),
		StorageKey:   request.StorageKey.String(),
		FileName:     request.FileName.String(),
		ContentType:  request.ContentType.String(),
		SizeBytes:    request.SizeBytes,
		ExpiresAt:    request.ExpiresAt.UTC().Format(time.RFC3339Nano),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	signature := directUploadSignature(payloadBytes, signingKey)
	return base64.RawURLEncoding.EncodeToString(payloadBytes) + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func decodeDirectUploadToken(token string, signingKey []byte) (ports.DirectAttachmentUploadRequest, error) {
	payloadPart, signaturePart, ok := strings.Cut(token, ".")
	if !ok {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	signature, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	if !hmac.Equal(signature, directUploadSignature(payloadBytes, signingKey)) {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}

	payload := directUploadTokenPayload{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	attachmentID, ok := media.NewID(payload.AttachmentID)
	if !ok {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	storageKey, ok := media.NewStorageKey(payload.StorageKey)
	if !ok {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	fileName, ok := media.NewFileName(payload.FileName)
	if !ok {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	contentType, ok := media.NewContentType(payload.ContentType)
	if !ok || payload.SizeBytes <= 0 {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, payload.ExpiresAt)
	if err != nil {
		return ports.DirectAttachmentUploadRequest{}, ports.ErrDirectUploadInvalid
	}
	return ports.DirectAttachmentUploadRequest{
		UploadID:     payload.UploadID,
		AttachmentID: attachmentID,
		TenantID:     tenant.ID(payload.TenantID),
		InventoryID:  inventory.InventoryID(payload.InventoryID),
		AssetID:      asset.ID(payload.AssetID),
		StorageKey:   storageKey,
		FileName:     fileName,
		ContentType:  contentType,
		SizeBytes:    payload.SizeBytes,
		ExpiresAt:    expiresAt,
	}, nil
}

func directUploadSignature(payload []byte, signingKey []byte) []byte {
	mac := hmac.New(sha256.New, signingKey)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

func copyStringMap(values map[string]string) map[string]string {
	copied := map[string]string{}
	for key, value := range values {
		copied[key] = value
	}
	return copied
}
