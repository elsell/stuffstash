package inputfiles

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"unicode/utf8"
)

func (Files) OpenUpload(ctx context.Context, path string) (ports.UploadFile, error) {
	if err := ctx.Err(); err != nil {
		return ports.UploadFile{}, err
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return ports.UploadFile{}, ports.Failure("file", "The CLI cannot read the upload file. Select a readable regular file.")
	}
	file, err := openInput(path)
	if err != nil {
		return ports.UploadFile{}, ports.Failure("file", "The CLI cannot open the upload file. Examine the path and file permissions.")
	}
	stop := context.AfterFunc(ctx, func() { file.Close() })
	keep := false
	defer func() {
		if !keep {
			stop()
			file.Close()
		}
	}()
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 {
		return ports.UploadFile{}, ports.Failure("file", "The upload file is empty or is not a regular file. Select another file.")
	}
	name := filepath.Base(path)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) > 255 {
		return ports.UploadFile{}, ports.Failure("file", "The file name is invalid or too long. Rename the file to use at most 255 characters.")
	}
	var prefix [512]byte
	n, err := file.ReadAt(prefix[:], 0)
	if ctx.Err() != nil {
		return ports.UploadFile{}, ctx.Err()
	}
	if err != nil && err != io.EOF {
		return ports.UploadFile{}, ports.Failure("file", "The CLI cannot read the upload file. Examine the file and try again.")
	}
	kind := http.DetectContentType(prefix[:n])
	switch kind {
	case "image/jpeg", "image/png", "image/webp", "application/pdf":
	default:
		return ports.UploadFile{}, ports.Failure("file", "This file type is not supported. Select a JPEG, PNG, WebP, or PDF file.")
	}
	keep = true
	return ports.UploadFile{Body: &uploadReader{ctx: ctx, file: file, stop: stop}, FileName: name, ContentType: kind, SizeBytes: info.Size()}, nil
}

type uploadReader struct {
	ctx  context.Context
	file *os.File
	stop func() bool
}

func (r *uploadReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := r.file.Read(p)
	if r.ctx.Err() != nil {
		return n, r.ctx.Err()
	}
	return n, err
}
func (r *uploadReader) Close() error { r.stop(); return r.file.Close() }
