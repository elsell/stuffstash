package push

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"net/http"
	"strings"
	"time"
)

const fcmTokenEndpoint = "https://oauth2.googleapis.com/token"
const fcmMessagingScope = "https://www.googleapis.com/auth/firebase.messaging"

type FCMAuth struct {
	config               *jwt.Config
	clock                ports.Clock
	client               *http.Client
	gate                 chan struct{}
	token                string
	expires, timeFetched time.Time
}

func NewFCMAuth(data []byte, clock ports.Clock, transport http.RoundTripper) (*FCMAuth, error) {
	if clock == nil || len(data) > 65536 {
		return nil, errFCM
	}
	var raw struct {
		Type     string `json:"type"`
		TokenURI string `json:"token_uri"`
	}
	if json.Unmarshal(data, &raw) != nil || raw.Type != "service_account" || (raw.TokenURI != "" && raw.TokenURI != fcmTokenEndpoint) {
		return nil, errFCM
	}
	config, err := google.JWTConfigFromJSON(data, fcmMessagingScope)
	if err != nil || config.Email == "" {
		return nil, errFCM
	}
	block, _ := pem.Decode(config.PrivateKey)
	if block == nil {
		return nil, errFCM
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errFCM
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok || key.N.BitLen() < 2048 || key.Validate() != nil {
		return nil, errFCM
	}
	config.TokenURL = fcmTokenEndpoint
	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	return &FCMAuth{config: config, clock: clock, gate: make(chan struct{}, 1), client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (a *FCMAuth) Token(ctx context.Context) (string, error) {
	select {
	case a.gate <- struct{}{}:
		defer func() { <-a.gate }()
	case <-ctx.Done():
		return "", ctx.Err()
	}
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	now := a.clock.Now()
	if a.token != "" && !now.Before(a.timeFetched) && now.Before(a.expires.Add(-time.Minute)) {
		return a.token, nil
	}
	a.token = ""
	client := *a.client
	client.Transport = fcmRefreshTransport{ctx: ctx, base: a.client.Transport}
	token, err := a.config.TokenSource(context.WithValue(ctx, oauth2.HTTPClient, &client)).Token()
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", errFCM
	}
	if token == nil || token.AccessToken == "" || len(token.AccessToken) > 16384 || strings.ContainsAny(token.AccessToken, "\r\n") || !strings.EqualFold(token.TokenType, "Bearer") || !token.Expiry.After(a.clock.Now()) {
		return "", errFCM
	}
	a.token = token.AccessToken
	a.expires = token.Expiry
	a.timeFetched = now
	return a.token, nil
}

// The pinned JWT source uses PostForm without a request context. Bind each
// refresh request to its caller without sharing a cancelled client between calls.
type fcmRefreshTransport struct {
	ctx  context.Context
	base http.RoundTripper
}

func (t fcmRefreshTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return t.base.RoundTrip(request.Clone(t.ctx))
}
