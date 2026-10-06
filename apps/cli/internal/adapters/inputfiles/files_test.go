package inputfiles

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRequestInputPreservesBytesAndRejectsUnsafeSources(t *testing.T) {
	ctx := context.Background()
	body := `{"name":null,"other":false}`
	files := Files{Stdin: strings.NewReader(body)}
	got, err := files.Read(ctx, "-")
	if err != nil || string(got) != body {
		t.Fatalf("stdin changed: %q %v", got, err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "request.json")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = files.Read(ctx, path)
	if err != nil || string(got) != body {
		t.Fatalf("file changed: %q %v", got, err)
	}
	for _, f := range []Files{{Stdin: strings.NewReader(body), StdinTerminal: true}, {Stdin: strings.NewReader(strings.Repeat("x", (1<<20)+1))}} {
		if _, err := f.Read(ctx, "-"); err == nil {
			t.Fatal("unsafe source accepted")
		}
	}
	if _, err := files.Read(ctx, dir); err == nil {
		t.Fatal("directory accepted")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := files.Read(canceled, path); err == nil {
		t.Fatal("canceled read accepted")
	}
}

type observedReader struct {
	*os.File
	started chan struct{}
}

func (r observedReader) Read(p []byte) (int, error) {
	select {
	case r.started <- struct{}{}:
	default:
	}
	return r.File.Read(p)
}

func TestCanceledPipeInputStopsReading(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	started := make(chan struct{}, 1)
	go func() { _, err := (Files{Stdin: observedReader{reader, started}}).Read(ctx, "-"); done <- err }()
	<-started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancel result: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled pipe input kept waiting")
	}
}
