package inventoryarchive

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestArchiveRejectsPathAliasesAndSpecialFiles(t *testing.T) {
	var valid bytes.Buffer
	if err := Write(context.Background(), &valid, []byte(`{"assets":[]}`), Selection{}, testTime(), nil, nil, limits()); err != nil {
		t.Fatal(err)
	}
	names := []string{
		"../inventory.json", "media/../../inventory.json", "/inventory.json",
		`..\inventory.json`, `C:\inventory.json`, `C:inventory.json`, `\\server\share\inventory.json`,
		"./inventory.json", "media/../inventory.json", "media//" + strings.Repeat("a", 64),
		"media/" + strings.Repeat("A", 64), "%2e%2e%2finventory.json", "inventory.json\x00ignored",
		"inventory.json:stream", "inventory.json.", "inventory.json ", "INVENTORY.JSON",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			data := rewriteInventoryEntry(t, valid.Bytes(), name, 0600)
			if _, err := Read(context.Background(), bytes.NewReader(data), int64(len(data)), limits()); err == nil {
				t.Fatal("unsafe ZIP path accepted")
			}
		})
	}
	for _, mode := range []os.FileMode{os.ModeSymlink | 0777, os.ModeDir | 0700, os.ModeNamedPipe | 0600, os.ModeDevice | os.ModeCharDevice | 0600} {
		t.Run(mode.String(), func(t *testing.T) {
			data := rewriteInventoryEntry(t, valid.Bytes(), inventoryName, mode)
			if _, err := Read(context.Background(), bytes.NewReader(data), int64(len(data)), limits()); err == nil {
				t.Fatal("special ZIP entry accepted")
			}
		})
	}
	// A benign central directory cannot conceal an unsafe local-header filename.
	data := append([]byte(nil), valid.Bytes()...)
	at := bytes.Index(data, []byte(inventoryName))
	if at < 0 {
		t.Fatal("missing local name")
	}
	copy(data[at:at+len(inventoryName)], []byte("../escape.json"))
	if _, err := Read(context.Background(), bytes.NewReader(data), int64(len(data)), limits()); err == nil {
		t.Fatal("mismatched local path accepted")
	}
}

func rewriteInventoryEntry(t *testing.T, original []byte, name string, mode os.FileMode) []byte {
	t.Helper()
	source, err := zip.NewReader(bytes.NewReader(original), int64(len(original)))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, file := range source.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		header := &zip.FileHeader{Name: file.Name, Method: zip.Store}
		if file.Name == inventoryName {
			header.Name = name
			header.SetMode(mode)
		}
		entry, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = entry.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
