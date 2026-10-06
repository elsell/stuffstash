package uploadtransfer

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestStreamingUploadAndCredentialIsolation(t *testing.T) {
	for _, method := range []string{"POST", "PUT"} {
		t.Run(method, func(t *testing.T) {
			calls := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" || r.Header.Get("X-Request-ID") != "" {
					t.Error("credentials crossed storage boundary")
				}
				if r.Method != method || r.ContentLength <= 0 {
					t.Error("method or length lost")
				}
				var data []byte
				if method == "POST" {
					parts, err := r.MultipartReader()
					if err != nil {
						t.Fatal(err)
					}
					p, err := parts.NextPart()
					if err != nil {
						t.Fatal(err)
					}
					b, _ := io.ReadAll(p)
					if p.FormName() != "policy" || string(b) != "signed" {
						t.Fatal("policy lost")
					}
					p, err = parts.NextPart()
					if err != nil {
						t.Fatal(err)
					}
					if p.FormName() != "file" || p.FileName() != "photo.png" || p.Header.Get("Content-Type") != "image/png" {
						t.Error("file metadata lost")
					}
					data, _ = io.ReadAll(p)
				} else {
					data, _ = io.ReadAll(r.Body)
				}
				if string(data) != "abcdef" {
					t.Errorf("file changed: %q", data)
				}
				w.WriteHeader(204)
			}))
			defer server.Close()
			client := server.Client()
			client.Jar, _ = cookiejar.New(nil)
			u, _ := url.Parse(server.URL)
			client.Jar.SetCookies(u, []*http.Cookie{{Name: "session", Value: "secret"}})
			instructions := ports.DirectUpload{URL: server.URL, Method: method}
			if method == "PUT" {
				instructions.Headers = map[string]string{"Content-Type": "image/png"}
			}
			if method == "POST" {
				instructions.FormFields = map[string]string{"policy": "signed"}
			}
			transfer := Client{HTTP: client}
			if err := transfer.Send(context.Background(), instructions, "photo.png", "image/png", 6, strings.NewReader("abcdef")); err != nil {
				t.Fatal(err)
			}
			for _, header := range []string{"Authorization", "Cookie", "Proxy-Authorization", "Host", "Content-Length", "Transfer-Encoding"} {
				bad := instructions
				bad.Headers = map[string]string{header: "secret"}
				if err := transfer.Send(context.Background(), bad, "photo.png", "image/png", 6, strings.NewReader("abcdef")); err == nil {
					t.Fatal("unsafe header accepted")
				}
			}
			if calls != 1 {
				t.Fatal("invalid requests reached server")
			}
		})
	}
}
func TestUploadRedirectAndShortInput(t *testing.T) {
	targetCalls := 0
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { targetCalls++ }))
	defer target.Close()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Location", target.URL)
		w.WriteHeader(307)
	}))
	defer server.Close()
	c := Client{HTTP: server.Client()}
	instruction := ports.DirectUpload{URL: server.URL, Method: "PUT"}
	if err := c.Send(context.Background(), instruction, "file", "application/pdf", 4, strings.NewReader("data")); err == nil {
		t.Fatal("redirect succeeded")
	}
	if targetCalls != 0 {
		t.Fatal("redirect followed")
	}
	instruction.URL = "http://example.com/secret"
	if err := c.Send(context.Background(), instruction, "file", "application/pdf", 4, strings.NewReader("data")); err == nil {
		t.Fatal("insecure destination accepted")
	}
	instruction.URL = server.URL
	if err := c.Send(context.Background(), instruction, "file", "application/pdf", 4, strings.NewReader("x")); err == nil {
		t.Fatal("short file accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Send(ctx, instruction, "file", "application/pdf", 4, strings.NewReader("data")); err != context.Canceled {
		t.Fatalf("cancellation lost: %v", err)
	}
}

// earlyStorage models a storage rejection while the transport still owns the body.
type earlyStorage struct{ finished chan struct{} }

func (s earlyStorage) RoundTrip(r *http.Request) (*http.Response, error) {
	go func() { defer close(s.finished); io.Copy(io.Discard, r.Body); r.Body.Close() }()
	return &http.Response{StatusCode: 403, Body: io.NopCloser(strings.NewReader("rejected")), Header: make(http.Header)}, nil
}
func TestEarlyStorageRejectionDuringBodyConsumption(t *testing.T) {
	finished := make(chan struct{})
	c := Client{HTTP: &http.Client{Transport: earlyStorage{finished: finished}}}
	err := c.Send(context.Background(), ports.DirectUpload{URL: "https://storage.example/upload", Method: "PUT"}, "file", "application/pdf", 1<<20, strings.NewReader(strings.Repeat("x", 1<<20)))
	<-finished
	failure, ok := err.(*ports.Error)
	if !ok || failure.Category != "api" {
		t.Fatalf("early rejection misclassified: %v", err)
	}
}
