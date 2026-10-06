package binaryfiles

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamOutput(t *testing.T) {
	var out bytes.Buffer
	f := Files{Stdout: &out}
	if err := f.PublishContent(context.Background(), "-", ports.BinaryContent{Body: io.NopCloser(strings.NewReader("payload")), ContentLength: 7}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "payload" {
		t.Fatal(out.String())
	}
}
func TestPrivateNoOverwriteAndTruncation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "file")
	f := Files{}
	body := func(s string, n int64) ports.BinaryContent {
		return ports.BinaryContent{Body: io.NopCloser(strings.NewReader(s)), ContentLength: n}
	}
	if err := f.PublishContent(context.Background(), p, body("short", 10)); err == nil {
		t.Fatal("accepted truncated file")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("published incomplete file")
	}
	if err := f.PublishContent(context.Background(), p, body("good", 4)); err != nil {
		t.Fatal(err)
	}
	if err := f.PublishContent(context.Background(), p, body("bad!", 4)); err == nil {
		t.Fatal("overwrote destination")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("wrong recovery advice: %v", err)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "good" {
		t.Fatal(string(b))
	}
}
