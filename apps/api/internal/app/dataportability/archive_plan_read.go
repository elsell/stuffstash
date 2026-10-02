package dataportability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s ArchiveService) loadRestorePlan(ctx context.Context, job archivejob.Record) (ports.ArchiveRestorePlan, error) {
	empty := ports.ArchiveRestorePlan{}
	key, ok := media.NewStorageKey(job.PlanArtifactID)
	if !ok {
		return empty, ErrArchiveMetadata
	}
	stream, size, err := s.deps.Storage.OpenBlobStream(ctx, key)
	if err != nil {
		return empty, err
	}
	defer stream.Close()
	if size <= 0 || size > s.deps.MaxMetadataBytes {
		return empty, ports.ErrBlobStreamSize
	}
	data, err := io.ReadAll(io.LimitReader(archiveInputReader{ctx: ctx, reader: stream}, size+1))
	if err != nil {
		return empty, err
	}
	hash := sha256.Sum256(data)
	if int64(len(data)) != size || hex.EncodeToString(hash[:]) != job.PlanSHA256 {
		return empty, ErrArchiveMetadata
	}
	plan, err := s.deps.Plans.DecodePlan(ctx, data, int(s.deps.MaxMetadataBytes))
	if err != nil {
		return empty, err
	}
	if plan.Document.TenantID != job.TenantID || plan.Document.InventoryID != job.DestinationInventoryID {
		return empty, ErrArchiveMetadata
	}
	if job.DestinationName != "" {
		plan.Document.InventoryName = job.DestinationName
	}
	if err = ValidateArchiveDocument(ctx, plan.Document, s.deps.MaxRecords); err != nil {
		return empty, err
	}
	return plan, nil
}
