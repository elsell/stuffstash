package dataportability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s ArchiveService) UploadRestore(ctx context.Context, a ArchiveAccess, key string, content io.Reader) (archivejob.Record, error) {
	empty := archivejob.Record{}
	if a.InventoryID != "" || content == nil || strings.TrimSpace(key) == "" {
		return empty, apperrors.ErrInvalidInput
	}
	if err := s.authorize(ctx, a); err != nil {
		return empty, err
	}
	scratch, err := s.deps.Scratch.NewArchiveScratch(ctx)
	if err != nil {
		return empty, err
	}
	defer scratch.Close()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(scratch, hash), io.LimitReader(archiveInputReader{ctx: ctx, reader: content}, s.deps.MaxArchiveBytes+1))
	if err != nil {
		return empty, err
	}
	if size <= 0 || size > s.deps.MaxArchiveBytes {
		return empty, ports.ErrBlobStreamSize
	}
	now := s.deps.Clock.Now()
	id := s.deps.IDs.NewID()
	artifact, ok := media.NewStorageKey("archives/" + a.TenantID.String() + "/" + id + "/source.zip")
	if !ok {
		return empty, apperrors.ErrInvalidInput
	}
	job, err := archivejob.New(archivejob.Request{RequestKey: key, ID: id, TenantID: a.TenantID.String(), PrincipalID: a.Principal.ID.String(), Kind: archivejob.Restore, SourceArtifactID: artifact.String(), SourceSHA256: hex.EncodeToString(hash.Sum(nil))}, now, now.Add(s.deps.Retention))
	if err != nil {
		return empty, err
	}
	// Publication of the queued job follows the complete private upload. A worker
	// can never claim a job whose source is still being written.
	err = s.deps.Storage.PutBlobStream(ctx, ports.BlobStreamWrite{Key: artifact, ContentType: "application/zip", SizeBytes: size, MaxBytes: s.deps.MaxArchiveBytes, Content: scratch})
	if err != nil {
		s.removeRedundantUpload(ctx, artifact)
		return empty, err
	}
	if err = s.authorize(ctx, a); err != nil {
		s.removeRedundantUpload(ctx, artifact)
		return empty, err
	}
	result, err := s.create(ctx, job)
	// A failed acknowledgment does not prove rollback. Retain the source until
	// retention cleanup can establish that no committed job references it.
	if err == nil && result.SourceArtifactID != artifact.String() {
		s.removeRedundantUpload(ctx, artifact)
	}
	return result, err
}
func (s ArchiveService) removeRedundantUpload(ctx context.Context, key media.StorageKey) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.deps.CleanupTimeout)
	defer cancel()
	// Retention cleanup retries deletion if storage is temporarily unavailable.
	if err := s.deps.Storage.DeleteBlob(cleanup, key); err != nil && s.deps.Observer != nil {
		s.deps.Observer.Record(cleanup, ports.Event{Name: ports.EventArchiveArtifactCleanupFailed, Message: "archive artifact cleanup deferred", Fields: map[string]string{"reason": "storage_unavailable"}})
	}
}

type archiveInputReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r archiveInputReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
