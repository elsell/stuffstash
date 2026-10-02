package inventoryarchive

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"
)

// Write streams a package into private staging storage. The caller must not
// publish that object unless Write succeeds; any error can leave partial bytes.
func Write(ctx context.Context, w io.Writer, inventory []byte, selection Selection, exportedAt time.Time, media []Media, source Source, limits Limits) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !limits.valid() {
		return ErrLimit
	}
	if !json.Valid(inventory) || exportedAt.IsZero() {
		return ErrInvalid
	}
	if int64(len(inventory)) > limits.MetadataBytes || int64(len(inventory)) > limits.EntryBytes {
		return ErrLimit
	}
	m := manifest{Format: format, Version: version, ExportedAt: exportedAt.UTC(), Selection: &selectionRecord{Photos: &selection.Photos, OtherFiles: &selection.OtherFiles}, Inventory: Media{hashBytes(inventory), int64(len(inventory))}, Media: []Media{}}
	seen := map[string]int64{}
	remaining := limits.ExpandedBytes - int64(len(inventory))
	if remaining < 0 {
		return ErrLimit
	}
	for _, entry := range media {
		if !validMedia(entry) {
			return ErrInvalid
		}
		if n, ok := seen[entry.SHA256]; ok {
			if n != entry.SizeBytes {
				return ErrInvalid
			}
			continue
		}
		if entry.SizeBytes > limits.EntryBytes || entry.SizeBytes > remaining {
			return ErrLimit
		}
		remaining -= entry.SizeBytes
		if len(seen) >= limits.Entries-2 {
			return ErrLimit
		}
		seen[entry.SHA256] = entry.SizeBytes
		m.Media = append(m.Media, entry)
	}
	if len(m.Media) > 0 && source == nil {
		return ErrInvalid
	}
	if len(m.Media) > limits.Entries-2 {
		return ErrLimit
	}
	// Bound the directory as well as manifest data so every successful write
	// fits the reader's allocation limit. Reserve the largest ZIP64 size/offset
	// extension for each header; ordinary ZIP headers are smaller.
	directoryBudget := limits.MetadataBytes
	for _, name := range []string{manifestName, inventoryName} {
		directoryBudget -= int64(46 + len(name) + 28)
	}
	if directoryBudget < 0 || int64(len(m.Media)) > directoryBudget/int64(46+len("media/")+64+28) {
		return ErrLimit
	}
	metadata, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if int64(len(metadata)) > remaining || int64(len(metadata)) > limits.MetadataBytes || int64(len(metadata)) > limits.EntryBytes {
		return ErrLimit
	}
	z := zip.NewWriter(&boundedWriter{ctx: ctx, w: w, remaining: limits.CompressedBytes})
	for _, entry := range []struct {
		name string
		data []byte
	}{{manifestName, metadata}, {inventoryName, inventory}} {
		f, err := z.Create(entry.name)
		if err != nil {
			return err
		}
		if _, err = io.Copy(f, &contextReader{ctx: ctx, r: bytes.NewReader(entry.data)}); err != nil {
			return err
		}
	}
	for _, entry := range m.Media {
		if err = ctx.Err(); err != nil {
			return err
		}
		r, err := source.Open(ctx, entry.SHA256)
		if err != nil {
			return err
		}
		f, err := z.Create("media/" + entry.SHA256)
		if err != nil {
			r.Close()
			return err
		}
		err = copyVerified(ctx, f, r, entry)
		closeErr := r.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return z.Close()
}

type boundedWriter struct {
	ctx       context.Context
	w         io.Writer
	remaining int64
}

func (w *boundedWriter) Write(b []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	if int64(len(b)) > w.remaining {
		return 0, ErrLimit
	}
	n, err := w.w.Write(b)
	w.remaining -= int64(n)
	return n, err
}
