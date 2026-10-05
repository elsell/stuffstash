package oidcauth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"golang.org/x/oauth2"
)

type Adapter struct {
	HTTP              *http.Client
	Clock             ports.Clock
	Output            ports.Output
	Browser           ports.Browser
	AllowLoopbackHTTP bool
}
type endpoints struct {
	Device string `json:"device_authorization_endpoint"`
	JWKS   string `json:"jwks_uri"`
}

func ValidateURL(raw string, allowLoopback bool) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("a valid HTTPS URL without credentials, query or fragment is required")
	}
	if u.Scheme == "https" {
		return nil
	}
	if allowLoopback && u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost" || u.Hostname() == "::1") {
		return nil
	}
	return errors.New("HTTPS is required; loopback HTTP needs explicit development configuration")
}
func (a Adapter) provider(ctx context.Context, issuer, clientID string) (context.Context, *oidc.Provider, oauth2.Config, error) {
	if ValidateURL(issuer, a.AllowLoopbackHTTP) != nil || clientID == "" {
		return ctx, nil, oauth2.Config{}, ports.Failure("configuration", "invalid OIDC issuer or client ID")
	}
	safe := *a.HTTP
	safe.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	ctx = oidc.ClientContext(ctx, &safe)
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return ctx, nil, oauth2.Config{}, ports.Failure("authentication", "OIDC discovery failed")
	}
	var extra endpoints
	if provider.Claims(&extra) != nil {
		return ctx, nil, oauth2.Config{}, ports.Failure("authentication", "invalid OIDC metadata")
	}
	endpoint := provider.Endpoint()
	endpoint.DeviceAuthURL = extra.Device
	endpoint.AuthStyle = oauth2.AuthStyleInParams
	for _, endpointURL := range []string{endpoint.AuthURL, endpoint.TokenURL, extra.JWKS, extra.Device} {
		if endpointURL != "" && ValidateURL(endpointURL, a.AllowLoopbackHTTP) != nil {
			return ctx, nil, oauth2.Config{}, ports.Failure("authentication", "OIDC metadata contains an insecure endpoint")
		}
	}
	return ctx, provider, oauth2.Config{ClientID: clientID, Endpoint: endpoint}, nil
}
func (a Adapter) Login(ctx context.Context, server string, c ports.AuthConfig, device bool) (ports.Session, error) {
	method := "authorization_code"
	if device {
		method = "device_code"
	}
	if !slices.Contains(c.LoginMethods, method) {
		return ports.Session{}, ports.Failure("unsupported", "requested login method is unavailable; check the CLI client configuration")
	}
	ctx, provider, config, err := a.provider(ctx, c.Issuer, c.ClientID)
	if err != nil {
		return ports.Session{}, err
	}
	config.Scopes = c.Scopes
	if !slices.Contains(c.Scopes, "openid") {
		return ports.Session{}, ports.Failure("configuration", "CLI login requires the openid scope")
	}
	var token *oauth2.Token
	nonce := ""
	if device {
		token, err = a.device(ctx, config)
	} else {
		token, nonce, err = a.browser(ctx, config, c)
	}
	if err != nil {
		return ports.Session{}, err
	}
	return a.session(ctx, provider, config.ClientID, server, token, nonce)
}
func (a Adapter) session(ctx context.Context, p *oidc.Provider, clientID, server string, token *oauth2.Token, nonce string) (ports.Session, error) {
	raw, _ := token.Extra("id_token").(string)
	if raw == "" {
		return ports.Session{}, ports.Failure("authentication", "provider did not return an OIDC ID token; configure openid support for this grant")
	}
	verified, err := p.Verifier(&oidc.Config{ClientID: clientID, Now: a.Clock.Now}).Verify(ctx, raw)
	if err != nil || verified.Subject == "" || nonce != "" && verified.Nonce != nonce {
		return ports.Session{}, ports.Failure("authentication", "provider identity token failed verification")
	}
	return ports.Session{Server: server, Issuer: verified.Issuer, Subject: verified.Subject, ClientID: clientID, IDToken: raw, RefreshToken: token.RefreshToken, ExpiresAt: verified.Expiry}, nil
}
func (a Adapter) Refresh(ctx context.Context, s ports.Session) (ports.Session, error) {
	if s.RefreshToken == "" {
		return ports.Session{}, ports.ErrNotLoggedIn
	}
	ctx, p, config, err := a.provider(ctx, s.Issuer, s.ClientID)
	if err != nil {
		return ports.Session{}, err
	}
	token, err := config.TokenSource(ctx, &oauth2.Token{RefreshToken: s.RefreshToken, Expiry: time.Unix(1, 0)}).Token()
	if err != nil {
		return ports.Session{}, ports.Failure("authentication", "session refresh failed; log in again")
	}
	result, err := a.session(ctx, p, s.ClientID, s.Server, token, "")
	if err == nil && (result.Issuer != s.Issuer || s.Subject != "" && result.Subject != s.Subject) {
		return ports.Session{}, ports.Failure("authentication", "The account changed during sign-in renewal. Log in again.")
	}
	if err == nil && result.RefreshToken == "" {
		result.RefreshToken = s.RefreshToken
	}
	return result, err
}
func (a Adapter) device(ctx context.Context, c oauth2.Config) (*oauth2.Token, error) {
	if c.Endpoint.DeviceAuthURL == "" {
		return nil, ports.Failure("unsupported", "provider does not advertise device authorization; use browser login where available")
	}
	response, err := c.DeviceAuth(ctx)
	if err != nil {
		return nil, ports.Failure("authentication", "could not begin device authorization")
	}
	defer func() { response.DeviceCode = "" }()
	if response.DeviceCode == "" || response.UserCode == "" || !response.Expiry.After(a.Clock.Now()) || ValidateURL(response.VerificationURI, a.AllowLoopbackHTTP) != nil {
		return nil, ports.Failure("authentication", "invalid device authorization response")
	}
	if err := a.Output.Notice("Open " + response.VerificationURI + " and enter code " + response.UserCode); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithDeadline(ctx, response.Expiry)
	defer cancel()
	for attempt := 0; attempt < 3; attempt++ {
		token, err := c.DeviceAccessToken(ctx, response)
		if err == nil {
			return token, nil
		}
		if ctx.Err() != nil {
			return nil, ports.Failure("authentication", "device authorization expired or was cancelled")
		}
		var rejection *oauth2.RetrieveError
		if errors.As(err, &rejection) {
			switch rejection.ErrorCode {
			case "access_denied":
				return nil, ports.Failure("authentication", "device authorization denied")
			case "expired_token":
				return nil, ports.Failure("authentication", "device authorization expired")
			}
			return nil, ports.Failure("authentication", "device authorization failed")
		}
		timer := time.NewTimer(time.Duration(attempt+1) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ports.Failure("authentication", "device authorization cancelled")
		case <-timer.C:
		}
	}
	return nil, ports.Failure("network", "device authorization connection failed; retry login")
}
func CanonicalServer(raw string) string { return strings.TrimRight(raw, "/") }
