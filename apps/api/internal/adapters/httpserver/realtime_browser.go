package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"nhooyr.io/websocket"
)

const browserRealtimeProtocol = "stuffstash.browser.v1"
const browserAuthenticationTimeout = 5 * time.Second
const browserAuthenticationMaxBytes = 16 * 1024

// acceptRealtime authenticates native headers or an explicitly negotiated browser
// frame. Neither path resolves providers or reads inventory before authentication.
func acceptRealtime(w http.ResponseWriter, r *http.Request, application app.App, allowedOrigins []string) (*websocket.Conn, identity.Principal, bool) {
	browser := false
	for _, protocol := range strings.Split(strings.Join(r.Header.Values("Sec-WebSocket-Protocol"), ","), ",") {
		if strings.TrimSpace(protocol) == browserRealtimeProtocol {
			browser = true
		}
	}
	origins := r.Header.Values("Origin")
	if len(origins) > 1 || (len(origins) == 1 && !allowedRealtimeOrigin(r, origins[0], allowedOrigins)) || (browser && len(origins) != 1) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil, identity.Principal{}, false
	}
	var principal identity.Principal
	var err error
	headerPresent := len(r.Header.Values("Authorization")) > 0
	if !browser || headerPresent {
		if len(r.Header.Values("Authorization")) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return nil, principal, false
		}
		principal, err = application.Authenticate(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return nil, principal, false
		}
	}
	options := &websocket.AcceptOptions{InsecureSkipVerify: true}
	if browser {
		options.Subprotocols = []string{browserRealtimeProtocol}
	}
	connection, err := websocket.Accept(w, r, options)
	if err != nil {
		return nil, principal, false
	}
	if browser && !headerPresent {
		connection.SetReadLimit(browserAuthenticationMaxBytes)
		ctx, cancel := context.WithTimeout(r.Context(), browserAuthenticationTimeout)
		defer cancel()
		kind, raw, readErr := connection.Read(ctx)
		var frame struct {
			Type          string `json:"type"`
			Authorization string `json:"authorization"`
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if readErr != nil || kind != websocket.MessageText || decoder.Decode(&frame) != nil || decoder.Decode(new(any)) != io.EOF || frame.Type != "session.authenticate" {
			_ = connection.Close(websocket.StatusPolicyViolation, "Authentication required")
			return nil, principal, false
		}
		principal, err = application.Authenticate(ctx, frame.Authorization)
		if err != nil {
			_ = connection.Close(websocket.StatusPolicyViolation, "Authentication required")
			return nil, principal, false
		}
		if connection.Write(ctx, websocket.MessageText, []byte(`{"type":"session.authenticated"}`)) != nil {
			_ = connection.CloseNow()
			return nil, principal, false
		}
	}
	return connection, principal, true
}

func allowedRealtimeOrigin(r *http.Request, origin string, allowed []string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return false
	}
	// Only the actual connection establishes same-origin scheme. TLS-terminating
	// deployments explicitly allow their public origin; forwarded headers are not trusted.
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if parsed.Scheme == scheme && strings.EqualFold(parsed.Host, r.Host) {
		return true
	}
	for _, candidate := range allowed {
		if origin == candidate {
			return true
		}
	}
	return false
}
