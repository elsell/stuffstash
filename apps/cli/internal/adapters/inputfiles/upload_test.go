package inputfiles

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadFilePreservesBytesAndMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "document.pdf")
	content := []byte("%PDF-1.7\nexample\n")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	file, err := (Files{}).OpenUpload(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Body.Close()
	if file.FileName != "document.pdf" || file.ContentType != "application/pdf" || file.SizeBytes != int64(len(content)) {
		t.Fatalf("wrong metadata: %+v", file)
	}
	data, err := io.ReadAll(file.Body)
	if err != nil || string(data) != string(content) {
		t.Fatal("sniff consumed file content")
	}
	cancel()
	if _, err = file.Body.Read(make([]byte, 1)); err != context.Canceled {
		t.Fatalf("cancellation lost: %v", err)
	}
}
func TestUploadFileRejectsInvalidInput(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty")
	text := filepath.Join(dir, "pretend.png")
	os.WriteFile(empty, nil, 0600)
	os.WriteFile(text, []byte("plain text"), 0600)
	for _, path := range []string{dir, empty, text, filepath.Join(dir, "missing")} {
		file, err := (Files{}).OpenUpload(context.Background(), path)
		if err == nil {
			file.Body.Close()
			t.Fatalf("invalid file accepted: %s", path)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (Files{}).OpenUpload(ctx, empty); err != context.Canceled {
		t.Fatal("canceled open lost")
	}
}
