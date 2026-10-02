package inventoryarchive

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"
)

type contentSource map[string][]byte

func (s contentSource) Open(_ context.Context, hash string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s[hash])), nil
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func testTime() time.Time    { return time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC) }
func limits() Limits {
	return Limits{CompressedBytes: 1 << 20, ExpandedBytes: 1 << 20, MetadataBytes: 64 << 10, EntryBytes: 128 << 10, Entries: 10}
}
func TestArchiveRoundTripDeduplicatesContentAndPreservesInventory(t *testing.T) {
	data := []byte(`{"schemaVersion":1,"inventoryName":"Kitchen 日本","tags":[{"displayName":"Medicine"}]}`)
	photo := []byte("original image bytes")
	hash := digest(photo)
	var out bytes.Buffer
	err := Write(context.Background(), &out, data, Selection{Photos: true, OtherFiles: true}, testTime(), []Media{{SHA256: hash, SizeBytes: int64(len(photo))}, {SHA256: hash, SizeBytes: int64(len(photo))}}, contentSource{hash: photo}, limits())
	if err != nil {
		t.Fatal(err)
	}
	a, err := Read(context.Background(), bytes.NewReader(out.Bytes()), int64(out.Len()), limits())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.InventoryJSON, data) || len(a.Media) != 1 || !a.Selection.Photos || !a.Selection.OtherFiles {
		t.Fatalf("lost content: %#v", a)
	}
	r, err := a.Open(hash)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	actual, err := io.ReadAll(r)
	if err != nil || !bytes.Equal(actual, photo) {
		t.Fatalf("media mismatch: %v", err)
	}
}
func TestArchiveRejectsUntrustedPackages(t *testing.T) {
	inventory := []byte(`{"assets":[]}`)
	media := []byte("file")
	hash := digest(media)
	var valid bytes.Buffer
	if err := Write(context.Background(), &valid, inventory, Selection{}, testTime(), []Media{{SHA256: hash, SizeBytes: 4}}, contentSource{hash: media}, limits()); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"traversal", "duplicate", "unlisted", "missing", "corrupt", "version", "unknown-manifest", "inventory-hash", "symlink", "selection-missing", "media-missing", "size-missing"} {
		t.Run(kind, func(t *testing.T) {
			z, _ := zip.NewReader(bytes.NewReader(valid.Bytes()), int64(valid.Len()))
			var out bytes.Buffer
			w := zip.NewWriter(&out)
			for _, f := range z.File {
				if kind == "missing" && f.Name == "media/"+hash {
					continue
				}
				r, _ := f.Open()
				b, _ := io.ReadAll(r)
				r.Close()
				if f.Name == "manifest.json" {
					var m map[string]any
					json.Unmarshal(b, &m)
					if kind == "selection-missing" {
						delete(m, "selection")
					}
					if kind == "media-missing" {
						delete(m, "media")
					}
					if kind == "size-missing" {
						delete(m["inventory"].(map[string]any), "sizeBytes")
					}
					if kind == "version" {
						m["archiveVersion"] = 99
					}
					if kind == "unknown-manifest" {
						m["unexpected"] = "value"
					}
					if kind == "inventory-hash" {
						m["inventory"].(map[string]any)["sha256"] = hash
					}
					b, _ = json.Marshal(m)
				}
				if kind == "corrupt" && f.Name == "media/"+hash {
					b = []byte("evil")
				}
				header := &zip.FileHeader{Name: f.Name}
				if kind == "symlink" && f.Name == "inventory.json" {
					header.SetMode(os.ModeSymlink | 0o777)
				}
				entry, _ := w.CreateHeader(header)
				entry.Write(b)
			}
			extra := ""
			switch kind {
			case "traversal":
				extra = "../escape"
			case "duplicate":
				extra = "inventory.json"
			case "unlisted":
				extra = "other"
			}
			if extra != "" {
				h := &zip.FileHeader{Name: extra}
				if kind == "symlink" {
					h.SetMode(os.ModeSymlink | 0o777)
				}
				e, _ := w.CreateHeader(h)
				e.Write([]byte("x"))
			}
			w.Close()
			if _, err := Read(context.Background(), bytes.NewReader(out.Bytes()), int64(out.Len()), limits()); err == nil {
				t.Fatal("unsafe archive accepted")
			}
		})
	}
}
func TestArchiveLimitsAndCancellation(t *testing.T) {
	var out bytes.Buffer
	data := []byte(`{"assets":[]}`)
	if err := Write(context.Background(), &out, data, Selection{}, testTime(), nil, nil, limits()); err != nil {
		t.Fatal(err)
	}
	for _, adjust := range []func(*Limits){func(l *Limits) { l.CompressedBytes = 1 }, func(l *Limits) { l.ExpandedBytes = 1 }, func(l *Limits) { l.MetadataBytes = 1 }, func(l *Limits) { l.Entries = 1 }, func(l *Limits) { l.EntryBytes = 1 }} {
		l := limits()
		adjust(&l)
		if _, err := Read(context.Background(), bytes.NewReader(out.Bytes()), int64(out.Len()), l); err == nil {
			t.Fatal("limit ignored")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Read(ctx, bytes.NewReader(out.Bytes()), int64(out.Len()), limits()); err == nil {
		t.Fatal("cancel ignored")
	}
	bad := []byte("not the original")
	h := digest([]byte("original"))
	out.Reset()
	if err := Write(context.Background(), &out, data, Selection{}, testTime(), []Media{{SHA256: h, SizeBytes: int64(len(bad))}}, contentSource{h: bad}, limits()); err == nil {
		t.Fatal("corrupt source published")
	}
}

func TestArchiveRejectsUnindexedBytesAndFalseDirectoryCount(t *testing.T) {
	var out bytes.Buffer
	if err := Write(context.Background(), &out, []byte(`{"assets":[]}`), Selection{}, testTime(), nil, nil, limits()); err != nil {
		t.Fatal(err)
	}
	valid := out.Bytes()
	falseCount := append([]byte(nil), valid...)
	falseCount[len(falseCount)-14] = 0
	falseCount[len(falseCount)-12] = 0
	for _, data := range [][]byte{append(append([]byte(nil), valid...), 0), append([]byte("prefix"), valid...), append(append([]byte(nil), valid...), valid...), falseCount} {
		if _, err := Read(context.Background(), bytes.NewReader(data), int64(len(data)), limits()); err == nil {
			t.Fatal("unindexed bytes or false directory accepted")
		}
	}
}

func TestWriterCannotExceedReaderDirectoryBudget(t *testing.T) {
	l := limits()
	l.MetadataBytes = 10000
	l.Entries = 110
	var media []Media
	source := contentSource{}
	for i := 0; i < 100; i++ {
		b := []byte{byte(i)}
		h := digest(b)
		media = append(media, Media{SHA256: h, SizeBytes: 1})
		source[h] = b
	}
	var out bytes.Buffer
	if err := Write(context.Background(), &out, []byte(`{"assets":[]}`), Selection{}, testTime(), media, source, l); err != ErrLimit {
		t.Fatalf("expected directory limit, got %v", err)
	}
}
func TestRequiredManifestCollectionsAndEmptyMediaSizes(t *testing.T) {
	for _, emptyFile := range []bool{false, true} {
		var valid bytes.Buffer
		var files []Media
		source := contentSource{}
		if emptyFile {
			h := digest(nil)
			files = []Media{{SHA256: h, SizeBytes: 0}}
			source[h] = nil
		}
		if err := Write(context.Background(), &valid, []byte(`{"assets":[]}`), Selection{}, testTime(), files, source, limits()); err != nil {
			t.Fatal(err)
		}
		z, err := zip.NewReader(bytes.NewReader(valid.Bytes()), int64(valid.Len()))
		if err != nil {
			t.Fatal(err)
		}
		var broken bytes.Buffer
		w := zip.NewWriter(&broken)
		for _, f := range z.File {
			r, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(r)
			r.Close()
			if err != nil {
				t.Fatal(err)
			}
			if f.Name == manifestName {
				var m map[string]any
				if err = json.Unmarshal(data, &m); err != nil {
					t.Fatal(err)
				}
				if emptyFile {
					delete(m["media"].([]any)[0].(map[string]any), "sizeBytes")
				} else {
					delete(m, "media")
				}
				data, err = json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
			}
			entry, err := w.Create(f.Name)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = entry.Write(data); err != nil {
				t.Fatal(err)
			}
		}
		if err = w.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err = Read(context.Background(), bytes.NewReader(broken.Bytes()), int64(broken.Len()), limits()); err == nil {
			t.Fatal("missing required manifest field accepted")
		}
	}
}
