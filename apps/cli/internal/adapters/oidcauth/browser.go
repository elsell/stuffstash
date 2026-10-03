package oidcauth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/oauth2"
)

func randomValue() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}
func (a Adapter) browser(ctx context.Context, c oauth2.Config, metadata ports.AuthConfig) (*oauth2.Token, string, error) {
	if metadata.LoopbackHost != "127.0.0.1" || metadata.LoopbackPathPrefix != "/callback/" || !metadata.EphemeralPort {
		return nil, "", ports.Failure("configuration", "provider CLI loopback redirect policy is unsupported")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", ports.Failure("authentication", "could not listen for browser login")
	}
	defer listener.Close()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	state, nonce, verifier := randomValue(), randomValue(), oauth2.GenerateVerifier()
	path := metadata.LoopbackPathPrefix + randomValue()
	c.RedirectURL = "http://" + listener.Addr().String() + path
	result := make(chan string, 1)
	var accepted atomic.Bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		if r.Method != "GET" || r.URL.Path != path || r.Host != listener.Addr().String() || subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("state")), []byte(state)) != 1 {
			http.Error(w, "Invalid login response", 400)
			return
		}
		if !accepted.CompareAndSwap(false, true) {
			http.Error(w, "Login response already received", 409)
			return
		}
		code := r.URL.Query().Get("code")
		if r.URL.Query().Get("error") != "" {
			code = ""
		}
		select {
		case result <- code:
			io.WriteString(w, "Return to Stuff Stash CLI to finish signing in.")
		default:
			http.Error(w, "Login response already received", 409)
		}
	})
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	defer server.Close()
	go func() { _ = server.Serve(listener) }()
	authURL := c.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("nonce", nonce), oauth2.AccessTypeOffline)
	if a.Browser == nil || a.Browser.Open(authURL) != nil {
		return nil, "", ports.Failure("authentication", "could not open a browser; use --device-code on a headless host")
	}
	var code string
	select {
	case <-ctx.Done():
		return nil, "", ports.Failure("authentication", "browser login expired or was cancelled")
	case code = <-result:
	}
	if code == "" {
		return nil, "", ports.Failure("authentication", "browser authorization was denied")
	}
	token, err := c.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, "", ports.Failure("authentication", "browser token exchange failed")
	}
	return token, nonce, nil
}

type SystemBrowser struct{}

func (SystemBrowser) Open(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}
