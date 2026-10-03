package oidcauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type localBrowser struct{ visit func(string) error }

func (b localBrowser) Open(value string) error { return b.visit(value) }
func TestBrowserPKCERejectsForgedAndRepeatedCallbacks(t *testing.T) {
	challenge := ""
	issuer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		if base64.RawURLEncoding.EncodeToString(sum[:]) != challenge || r.Form.Get("code") != "authorized" {
			t.Error("PKCE exchange not bound to authorization")
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"access_token":"access","token_type":"Bearer","id_token":"identity"}`)
	}))
	defer issuer.Close()
	adapter := Adapter{Browser: localBrowser{visit: func(raw string) error {
		auth, _ := url.Parse(raw)
		values := auth.Query()
		challenge = values.Get("code_challenge")
		if challenge == "" || values.Get("code_challenge_method") != "S256" || values.Get("nonce") == "" {
			t.Fatal("authorization missing PKCE or nonce")
		}
		callback, _ := url.Parse(values.Get("redirect_uri"))
		if callback.Hostname() != "127.0.0.1" {
			t.Fatal("non-loopback callback")
		}
		for _, attempt := range []struct {
			state  string
			status int
		}{{"forged", 400}, {values.Get("state"), 200}, {values.Get("state"), 409}} {
			callback.RawQuery = url.Values{"state": {attempt.state}, "code": {"authorized"}}.Encode()
			response, err := http.Get(callback.String())
			if err != nil {
				return err
			}
			response.Body.Close()
			if response.StatusCode != attempt.status {
				t.Errorf("callback status %d want %d", response.StatusCode, attempt.status)
			}
		}
		return nil
	}}}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, issuer.Client())
	token, nonce, err := adapter.browser(ctx, oauth2.Config{ClientID: "cli", Endpoint: oauth2.Endpoint{AuthURL: issuer.URL + "/auth", TokenURL: issuer.URL + "/token", AuthStyle: oauth2.AuthStyleInParams}}, ports.AuthConfig{LoopbackHost: "127.0.0.1", LoopbackPathPrefix: "/callback/", EphemeralPort: true})
	if err != nil || nonce == "" || token == nil {
		t.Fatalf("browser login failed: %v", err)
	}
}
