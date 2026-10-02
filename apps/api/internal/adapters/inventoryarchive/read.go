package inventoryarchive

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
)

// Read validates the complete container and every checksum without extracting
// files. It retains only bounded metadata; media is checked by streaming.
func Read(ctx context.Context, source io.ReaderAt, size int64, limits Limits) (*Archive, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !limits.valid() || size < 0 || size > limits.CompressedBytes {
		return nil, ErrLimit
	}
	directory, err := preflight(ctx, source, size, limits)
	if err != nil {
		return nil, err
	}
	z, err := zip.NewReader(source, size)
	if err != nil {
		return nil, ErrInvalid
	}
	if len(z.File) > limits.Entries {
		return nil, ErrLimit
	}
	if err = verifyCoverage(ctx, source, z, directory); err != nil {
		return nil, err
	}
	entries := make(map[string]*zip.File, len(z.File))
	remaining := limits.ExpandedBytes
	for _, f := range z.File {
		if !allowedName(f.Name) || !f.Mode().IsRegular() || f.Flags&1 != 0 || entries[f.Name] != nil {
			return nil, ErrInvalid
		}
		if f.Method != zip.Store && f.Method != zip.Deflate {
			return nil, ErrInvalid
		}
		if f.UncompressedSize64 > uint64(limits.EntryBytes) || f.UncompressedSize64 > uint64(remaining) {
			return nil, ErrLimit
		}
		remaining -= int64(f.UncompressedSize64)
		entries[f.Name] = f
	}
	mf, ok := entries[manifestName]
	if !ok {
		return nil, ErrInvalid
	}
	mb, err := readMetadata(ctx, mf, limits.MetadataBytes)
	if err != nil {
		return nil, err
	}
	var m manifest
	decoder := json.NewDecoder(bytes.NewReader(mb))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&m) != nil || decoder.Decode(new(any)) != io.EOF {
		return nil, ErrInvalid
	}
	if m.Format != format || m.Version != version || m.ExportedAt.IsZero() || m.Media == nil || m.Selection == nil || m.Selection.Photos == nil || m.Selection.OtherFiles == nil || !validMedia(m.Inventory) {
		return nil, ErrInvalid
	}
	inv, ok := entries[inventoryName]
	if !ok {
		return nil, ErrInvalid
	}
	data, err := readMetadata(ctx, inv, limits.MetadataBytes)
	if err != nil {
		return nil, err
	}
	if !json.Valid(data) || int64(len(data)) != m.Inventory.SizeBytes || hashBytes(data) != m.Inventory.SHA256 {
		return nil, ErrInvalid
	}
	a := &Archive{InventoryJSON: data, Selection: Selection{*m.Selection.Photos, *m.Selection.OtherFiles}, ExportedAt: m.ExportedAt, Media: m.Media, files: map[string]*zip.File{}}
	if len(m.Media) != len(entries)-2 {
		return nil, ErrInvalid
	}
	for _, media := range m.Media {
		if !validMedia(media) || a.files[media.SHA256] != nil {
			return nil, ErrInvalid
		}
		f, ok := entries["media/"+media.SHA256]
		if !ok || f.UncompressedSize64 != uint64(media.SizeBytes) {
			return nil, ErrInvalid
		}
		r, err := f.Open()
		if err != nil {
			return nil, ErrInvalid
		}
		err = copyVerified(ctx, io.Discard, r, media)
		closeErr := r.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		a.files[media.SHA256] = f
	}
	return a, nil
}
func readMetadata(ctx context.Context, f *zip.File, max int64) ([]byte, error) {
	if f.UncompressedSize64 > uint64(max) {
		return nil, ErrLimit
	}
	r, err := f.Open()
	if err != nil {
		return nil, ErrInvalid
	}
	defer r.Close()
	var out bytes.Buffer
	if _, err = io.Copy(&out, &contextReader{ctx: ctx, r: io.LimitReader(r, max)}); err != nil {
		return nil, err
	}
	var extra [1]byte
	n, err := r.Read(extra[:])
	if n != 0 {
		return nil, ErrLimit
	}
	if err != io.EOF {
		return nil, ErrInvalid
	}
	if uint64(out.Len()) != f.UncompressedSize64 {
		return nil, ErrInvalid
	}
	return out.Bytes(), nil
}
func allowedName(name string) bool {
	return name == manifestName || name == inventoryName || (strings.HasPrefix(name, "media/") && validHash(strings.TrimPrefix(name, "media/")))
}
func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validMedia(m Media) bool   { return m.SizeBytes >= 0 && validHash(m.SHA256) }
func hashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(b []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(b)
}
func copyVerified(ctx context.Context, w io.Writer, r io.Reader, m Media) error {
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(w, h), &contextReader{ctx: ctx, r: io.LimitReader(r, m.SizeBytes)})
	if err != nil {
		return err
	}
	if n != m.SizeBytes {
		return ErrInvalid
	}
	var extra [1]byte
	count, err := (&contextReader{ctx: ctx, r: r}).Read(extra[:])
	if count != 0 {
		return ErrInvalid
	}
	if err != io.EOF {
		return ErrInvalid
	}
	if hex.EncodeToString(h.Sum(nil)) != m.SHA256 {
		return ErrInvalid
	}
	return ctx.Err()
}
