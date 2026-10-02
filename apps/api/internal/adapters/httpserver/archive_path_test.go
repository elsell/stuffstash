package httpserver

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func appendArchivePath(t *testing.T, original []byte, name string) []byte {
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
		entry, err := writer.Create(file.Name)
		if err != nil {
			reader.Close()
			t.Fatal(err)
		}
		_, err = io.Copy(entry, reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	entry, err := writer.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = entry.Write([]byte("must never be extracted")); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
