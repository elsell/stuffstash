package dataportability

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type archiveWorkResult struct {
	artifact, planHash, destination string
	plan                            *ports.ArchiveRestorePlan
	err                             error
}

func (w ArchiveWorker) execute(ctx context.Context, job archivejob.Record) archiveWorkResult {
	if job.Kind == archivejob.Export {
		id, err := w.export(ctx, job)
		return archiveWorkResult{artifact: id, err: err}
	}
	if job.Phase == archivejob.Validation {
		return w.previewRestore(ctx, job)
	}
	plan, err := w.stageRestore(ctx, job)
	return archiveWorkResult{plan: &plan, err: err}
}
func (w ArchiveWorker) openRestore(ctx context.Context, job archivejob.Record) (ports.ArchivePackageContents, io.Closer, error) {
	empty := ports.ArchivePackageContents{}
	key, ok := media.NewStorageKey(job.SourceArtifactID)
	if !ok {
		return empty, nil, ErrArchiveMetadata
	}
	stream, size, err := w.deps.Service.deps.Storage.OpenBlobStream(ctx, key)
	if err != nil {
		return empty, nil, err
	}
	fail := func(err error) (ports.ArchivePackageContents, io.Closer, error) {
		stream.Close()
		return empty, nil, err
	}
	if size <= 0 || size > w.deps.Limits.CompressedBytes {
		return fail(ports.ErrBlobStreamSize)
	}
	hash := sha256.New()
	n, err := io.Copy(hash, io.LimitReader(archiveInputReader{ctx: ctx, reader: stream}, size+1))
	if err != nil {
		return fail(err)
	}
	if n != size || hex.EncodeToString(hash.Sum(nil)) != job.SourceSHA256 {
		return fail(ErrArchiveMetadata)
	}
	archive, err := w.deps.Readers.ReadPackage(ctx, stream, size, w.deps.Limits)
	if err != nil {
		return fail(err)
	}
	return archive, stream, nil
}
func (w ArchiveWorker) previewRestore(ctx context.Context, job archivejob.Record) archiveWorkResult {
	failed := func(err error) archiveWorkResult { return archiveWorkResult{err: err} }
	archive, closer, err := w.openRestore(ctx, job)
	if err != nil {
		return failed(err)
	}
	defer closer.Close()
	doc, err := w.deps.Metadata.DecodeMetadata(ctx, archive.Metadata, int(w.deps.Limits.MetadataBytes))
	if err != nil {
		return failed(err)
	}
	if err = ValidateArchiveDocument(ctx, doc, w.deps.MaxRecords); err != nil {
		return failed(err)
	}
	if err = validateArchiveMediaSet(doc, archive); err != nil {
		return failed(err)
	}
	reservations, err := w.restoreReservations(ctx, job)
	if err != nil {
		return failed(err)
	}
	destination := w.deps.Service.deps.IDs.NewID()
	plan, err := BuildArchiveRestorePlan(ctx, doc, ports.ArchiveRestoreDestination{TenantID: job.TenantID, InventoryID: destination, Name: doc.InventoryName, PrincipalID: job.PrincipalID}, reservations, ports.ArchiveMediaSelection{Photos: archive.Photos, OtherFiles: archive.OtherFiles}, w.deps.Service.deps.IDs, w.deps.MaxRecords)
	if err != nil {
		return failed(err)
	}
	data, err := w.deps.Plans.EncodePlan(ctx, plan, int(w.deps.Limits.MetadataBytes))
	if err != nil {
		return failed(err)
	}
	hash := sha256.Sum256(data)
	key, ok := media.NewStorageKey("archives/" + job.TenantID + "/" + job.ID + "/" + job.LeaseToken + "/plan.json")
	if !ok {
		return failed(ErrArchiveMetadata)
	}
	if err = w.deps.Service.deps.Artifacts.RegisterArchiveArtifact(ctx, job, key, ports.ArchiveArtifactPlan, w.deps.Service.deps.Clock.Now()); err != nil {
		return failed(err)
	}
	if err = w.deps.Service.deps.Storage.PutBlobStream(ctx, ports.BlobStreamWrite{Key: key, ContentType: "application/json", SizeBytes: int64(len(data)), MaxBytes: w.deps.Limits.MetadataBytes, Content: bytes.NewReader(data)}); err != nil {
		return failed(err)
	}
	return archiveWorkResult{artifact: key.String(), planHash: hex.EncodeToString(hash[:]), destination: destination}
}
func validateArchiveMediaSet(doc ports.InventoryExportDocument, archive ports.ArchivePackageContents) error {
	if !doc.ExportedAt.Equal(archive.ExportedAt) {
		return ErrArchiveMetadata
	}
	expected := map[string]int64{}
	for _, a := range doc.Assets {
		for _, m := range a.Attachments {
			if m.ContentType.IsImage() && !archive.Photos || !m.ContentType.IsImage() && !archive.OtherFiles {
				continue
			}
			hash := m.SHA256.String()
			if n, exists := expected[hash]; exists && n != m.SizeBytes {
				return ErrArchiveMetadata
			}
			expected[hash] = m.SizeBytes
		}
	}
	if len(expected) != len(archive.Media) {
		return ErrArchiveMetadata
	}
	for _, m := range archive.Media {
		if expected[m.SHA256] != m.SizeBytes {
			return ErrArchiveMetadata
		}
		delete(expected, m.SHA256)
	}
	return nil
}
func (w ArchiveWorker) stageRestore(ctx context.Context, job archivejob.Record) (ports.ArchiveRestorePlan, error) {
	plan, err := w.deps.Service.loadRestorePlan(ctx, job)
	if err != nil {
		return plan, err
	}
	archive, closer, err := w.openRestore(ctx, job)
	if err != nil {
		return plan, err
	}
	defer closer.Close()
	for _, a := range plan.Document.Assets {
		for _, m := range a.Attachments {
			if err = w.stageRestoreMedia(ctx, job, archive.Source, m); err != nil {
				return plan, err
			}
		}
	}
	return plan, nil
}
func (w ArchiveWorker) stageRestoreMedia(ctx context.Context, job archivejob.Record, source ports.ArchiveContentSource, m media.Attachment) error {
	if m.SizeBytes <= 0 || m.SizeBytes > w.deps.Limits.EntryBytes {
		return ports.ErrBlobStreamSize
	}
	input, err := source.Open(ctx, m.SHA256.String())
	if err != nil {
		return err
	}
	defer input.Close()
	scratch, err := w.deps.Service.deps.Scratch.NewArchiveScratch(ctx)
	if err != nil {
		return err
	}
	defer scratch.Close()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(scratch, hash), io.LimitReader(archiveInputReader{ctx: ctx, reader: input}, m.SizeBytes+1))
	if err != nil {
		return err
	}
	if size != m.SizeBytes || hex.EncodeToString(hash.Sum(nil)) != m.SHA256.String() {
		return ErrArchiveMetadata
	}
	if err = w.deps.Service.deps.Artifacts.RegisterArchiveArtifact(ctx, job, m.StorageKey, ports.ArchiveArtifactRestoredMedia, w.deps.Service.deps.Clock.Now()); err != nil {
		return err
	}
	return w.deps.Service.deps.Storage.PutBlobStream(ctx, ports.BlobStreamWrite{Key: m.StorageKey, ContentType: m.ContentType.String(), SizeBytes: size, MaxBytes: w.deps.Limits.EntryBytes, Content: scratch})
}
