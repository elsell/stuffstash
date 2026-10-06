package uploadtransfer

import (
	"bytes"
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"math"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type Client struct {
	HTTP              *http.Client
	AllowLoopbackHTTP bool
}

func (c Client) Send(ctx context.Context, in ports.DirectUpload, name, contentType string, size int64, source io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.validate(in); err != nil {
		return err
	}
	if size <= 0 || source == nil || strings.ContainsAny(contentType, "\r\n") {
		return ports.Failure("file", "The file size or type is invalid. Choose a non-empty file with a valid media type.")
	}
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		return ports.Failure("file", "The file type is invalid. Supply a valid media type.")
	}
	for name, value := range in.Headers {
		if strings.EqualFold(name, "Content-Type") && (in.Method != "PUT" || value != contentType) {
			return ports.Failure("protocol", "The upload media type does not match the file. Start a new upload.")
		}
	}
	input := &io.LimitedReader{R: source, N: size}
	var body io.Reader = input
	length := size
	requestType := contentType
	if in.Method == "POST" {
		prefix, suffix, kind, err := multipartParts(in.FormFields, name, contentType)
		if err != nil {
			return err
		}
		overhead := int64(len(prefix)) + int64(len(suffix))
		if size > math.MaxInt64-overhead {
			return ports.Failure("file", "The file is too large. Choose a smaller file.")
		}
		length += overhead
		requestType = kind
		body = io.MultiReader(bytes.NewReader(prefix), input, bytes.NewReader(suffix))
	}
	tracked := &uploadBody{Reader: body, closed: make(chan struct{})}
	request, err := http.NewRequestWithContext(ctx, in.Method, in.URL, tracked)
	if err != nil {
		return ports.Failure("protocol", "The upload address is invalid. Start a new upload.")
	}
	request.ContentLength = length
	for name, value := range in.Headers {
		request.Header.Set(name, value)
	}
	request.Header.Set("Content-Type", requestType)
	if c.HTTP == nil {
		return ports.Failure("configuration", "File transfers are not configured. Update the CLI and try again.")
	}
	client := *c.HTTP
	client.Jar = nil
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return ports.Failure("network", "The file transfer failed. Check the upload status before you retry.")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ports.Failure("api", "Storage did not accept the file. Check the upload status before you retry.")
	}
	select {
	case <-tracked.closed:
	case <-ctx.Done():
		return ctx.Err()
	}
	if tracked.read.Load() != length {
		return ports.Failure("file", "The file transfer ended early. Start a new upload with the complete file.")
	}
	return nil
}
func (c Client) validate(in ports.DirectUpload) error {
	invalid := func() error {
		return ports.Failure("protocol", "The storage upload instructions are invalid. Start a new upload.")
	}
	u, err := url.Parse(in.URL)
	if err != nil || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return invalid()
	}
	local := u.Hostname() == "localhost" || net.ParseIP(u.Hostname()).IsLoopback()
	if u.Scheme != "https" && !(u.Scheme == "http" && c.AllowLoopbackHTTP && local) {
		return invalid()
	}
	if in.Method != "POST" && in.Method != "PUT" {
		return invalid()
	}
	if in.Method == "PUT" && len(in.FormFields) != 0 {
		return invalid()
	}
	for name := range in.Headers {
		switch strings.ToLower(name) {
		case "authorization", "proxy-authorization", "cookie", "host", "content-length", "transfer-encoding", "connection", "trailer":
			return invalid()
		}
	}
	if _, exists := in.FormFields["file"]; exists {
		return invalid()
	}
	return nil
}
func multipartParts(fields map[string]string, name, contentType string) ([]byte, []byte, string, error) {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := writer.WriteField(key, fields[key]); err != nil {
			return nil, nil, "", err
		}
	}
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "file", "filename": name}))
	header.Set("Content-Type", contentType)
	if _, err := writer.CreatePart(header); err != nil {
		return nil, nil, "", err
	}
	prefix := append([]byte(nil), b.Bytes()...)
	b.Reset()
	if err := writer.Close(); err != nil {
		return nil, nil, "", err
	}
	return prefix, append([]byte(nil), b.Bytes()...), writer.FormDataContentType(), nil
}

// The HTTP transport can consume and close a request body after Do returns.
type uploadBody struct {
	io.Reader
	read   atomic.Int64
	closed chan struct{}
	once   sync.Once
}

func (b *uploadBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	b.read.Add(int64(n))
	return n, err
}
func (b *uploadBody) Close() error { b.once.Do(func() { close(b.closed) }); return nil }
