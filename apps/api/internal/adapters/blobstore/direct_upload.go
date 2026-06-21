package blobstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type LocalDirectAttachmentUploader struct {
	blobs   ports.BlobStorage
	mu      sync.Mutex
	pending map[string]ports.DirectAttachmentUploadRequest
}

func NewLocalDirectAttachmentUploader(blobs ports.BlobStorage) *LocalDirectAttachmentUploader {
	return &LocalDirectAttachmentUploader{
		blobs:   blobs,
		pending: map[string]ports.DirectAttachmentUploadRequest{},
	}
}

func (u *LocalDirectAttachmentUploader) CreateDirectAttachmentUpload(_ context.Context, request ports.DirectAttachmentUploadRequest) (ports.DirectAttachmentUpload, error) {
	if u == nil || u.blobs == nil || strings.TrimSpace(request.UploadID) == "" || request.ExpiresAt.IsZero() || !request.ExpiresAt.After(time.Now().UTC()) {
		return ports.DirectAttachmentUpload{}, ports.ErrDirectUploadInvalid
	}

	u.mu.Lock()
	u.pending[request.UploadID] = request
	u.mu.Unlock()

	return ports.DirectAttachmentUpload{
		UploadID:     request.UploadID,
		AttachmentID: request.AttachmentID,
		Method:       "PUT",
		URL:          "stuffstash-local://direct-uploads/" + request.UploadID,
		Headers:      map[string]string{"content-type": request.ContentType.String()},
		ExpiresAt:    request.ExpiresAt,
	}, nil
}

func (u *LocalDirectAttachmentUploader) CompleteDirectAttachmentUpload(ctx context.Context, uploadID string) (ports.CompletedDirectAttachmentUpload, error) {
	if u == nil || u.blobs == nil {
		return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadInvalid
	}

	u.mu.Lock()
	request, ok := u.pending[uploadID]
	if ok {
		delete(u.pending, uploadID)
	}
	u.mu.Unlock()
	if !ok {
		return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadIncomplete
	}
	if !request.ExpiresAt.After(time.Now().UTC()) {
		return ports.CompletedDirectAttachmentUpload{}, ports.ErrDirectUploadExpired
	}

	data, err := u.blobs.GetBlob(ctx, request.StorageKey)
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
