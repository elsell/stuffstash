package dataportability

import (
	"context"
	"io"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Package first, publish later: this function cannot mark the job ready.
func (w ArchiveWorker) export(ctx context.Context, job archivejob.Record) (string, error) {
	s := w.deps.Service
	doc, err := CaptureArchiveMetadata(ctx, w.deps.Snapshots, tenant.ID(job.TenantID), inventory.InventoryID(job.SourceInventoryID), s.deps.Clock.Now(), w.deps.MaxRecords)
	if err != nil {
		return "", err
	}
	metadata, err := w.deps.Metadata.EncodeMetadata(ctx, doc, int(w.deps.Limits.MetadataBytes))
	if err != nil {
		return "", err
	}
	input := ports.ArchivePackageInput{Metadata: metadata, Photos: job.Photos, OtherFiles: job.OtherFiles, ExportedAt: doc.ExportedAt}
	source := archiveExportSource{storage: s.deps.Storage, attachments: map[string]media.Attachment{}}
	for _, a := range doc.Assets {
		for _, m := range a.Attachments {
			if m.ContentType.IsImage() && !job.Photos || !m.ContentType.IsImage() && !job.OtherFiles {
				continue
			}
			hash := m.SHA256.String()
			if prior, ok := source.attachments[hash]; ok {
				if prior.SizeBytes != m.SizeBytes {
					return "", ErrArchiveMetadata
				}
				continue
			}
			source.attachments[hash] = m
			input.Media = append(input.Media, ports.ArchivePackageMedia{SHA256: hash, SizeBytes: m.SizeBytes})
		}
	}
	scratch, err := s.deps.Scratch.NewArchiveScratch(ctx)
	if err != nil {
		return "", err
	}
	defer scratch.Close()
	if err = w.deps.Packages.WritePackage(ctx, scratch, input, source, w.deps.Limits); err != nil {
		return "", err
	}
	size, err := scratch.Seek(0, io.SeekEnd)
	if err != nil {
		return "", err
	}
	key, ok := media.NewStorageKey("archives/" + job.TenantID + "/" + job.ID + "/" + job.LeaseToken + "/export.zip")
	if !ok {
		return "", archivejob.ErrInvalid
	}
	if err = s.deps.Artifacts.RegisterArchiveArtifact(ctx, job, key, ports.ArchiveArtifactExport, s.deps.Clock.Now()); err != nil {
		return "", err
	}
	if err = s.deps.Storage.PutBlobStream(ctx, ports.BlobStreamWrite{Key: key, ContentType: "application/zip", SizeBytes: size, MaxBytes: w.deps.Limits.CompressedBytes, Content: scratch}); err != nil {
		return "", err
	}
	return key.String(), nil
}

type archiveExportSource struct {
	storage     ports.StreamingBlobStorage
	attachments map[string]media.Attachment
}

func (s archiveExportSource) Open(ctx context.Context, hash string) (io.ReadCloser, error) {
	attachment, found := s.attachments[hash]
	if !found {
		return nil, ErrArchiveMetadata
	}
	stream, size, err := s.storage.OpenBlobStream(ctx, attachment.StorageKey)
	if err != nil {
		return nil, err
	}
	if size != attachment.SizeBytes {
		stream.Close()
		return nil, ports.ErrBlobStreamSize
	}
	return stream, nil
}
