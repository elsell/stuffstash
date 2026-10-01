// Package mcpserver hosts authenticated MCP over existing application services.
package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const Path = "/mcp"
const MetadataPath = "/.well-known/oauth-protected-resource/mcp"

type Options struct {
	PublicURL      string
	AuthMode       string
	Issuer         string
	AllowedOrigins []string
	MaxBodyBytes   int64
	Observer       ports.Observer
	Clock          ports.Clock
	RequestTimeout time.Duration
}

type handler struct {
	application app.App
	options     Options
	endpoint    *url.URL
}

func New(application app.App, options Options) (http.Handler, error) {
	endpoint, err := url.Parse(options.PublicURL)
	if err != nil || endpoint.Host == "" || endpoint.EscapedPath() != Path || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.ForceQuery || endpoint.Fragment != "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") {
		return nil, fmt.Errorf("invalid MCP public endpoint")
	}
	if options.AuthMode != "local-dev" && options.AuthMode != "oidc" {
		return nil, fmt.Errorf("MCP authentication mode must be explicit")
	}
	if options.AuthMode == "oidc" && (options.Issuer == "" || endpoint.Scheme != "https") {
		return nil, fmt.Errorf("MCP OIDC endpoint requires HTTPS and an issuer")
	}
	if endpoint.Scheme == "http" {
		ip := net.ParseIP(endpoint.Hostname())
		if options.AuthMode != "local-dev" || (endpoint.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback())) {
			return nil, fmt.Errorf("MCP HTTP requires local-dev loopback")
		}
	}
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = 30 * time.Second
	}
	if options.MaxBodyBytes <= 0 {
		options.MaxBodyBytes = 1024 * 1024
	}
	if options.Clock == nil {
		options.Clock = ports.SystemClock{}
	}
	return &handler{application: application, options: options, endpoint: endpoint}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.options.RequestTimeout)
	defer cancel()
	r = r.WithContext(ctx)
	w.Header().Set("Cache-Control", "private, no-store")
	if !strings.EqualFold(r.Host, h.endpoint.Host) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	origins := r.Header.Values("Origin")
	if len(origins) > 1 || (len(origins) == 1 && !h.allowedOrigin(origins[0])) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.URL.Path == MetadataPath {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", 405)
			return
		}
		issuers := []string{}
		if h.options.Issuer != "" {
			issuers = append(issuers, h.options.Issuer)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"resource": h.options.PublicURL, "authorization_servers": issuers, "bearer_methods_supported": []string{"header"}})
		return
	}
	if r.URL.Path != Path {
		http.NotFound(w, r)
		return
	}
	authorization := r.Header.Get("Authorization")
	if len(r.Header.Values("Authorization")) != 1 || (h.options.AuthMode == "oidc" && strings.HasPrefix(strings.ToLower(authorization), "bearer dev:")) {
		h.unauthorized(w)
		return
	}
	principal, err := h.application.Authenticate(r.Context(), authorization)
	if err != nil {
		h.unauthorized(w)
		return
	}
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, h.options.MaxBodyBytes))
		if err != nil {
			http.Error(w, "Request body unavailable or too large", http.StatusRequestEntityTooLarge)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
	}
	transport := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return h.server(r.Context(), principal) }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true, MaxRequestBodyBytes: h.options.MaxBodyBytes})
	transport.ServeHTTP(w, r)
}

func (h *handler) allowedOrigin(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.Path != "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return false
	}
	if value == h.endpoint.Scheme+"://"+h.endpoint.Host {
		return true
	}
	for _, allowed := range h.options.AllowedOrigins {
		if value == allowed {
			return true
		}
	}
	return false
}
func (h *handler) unauthorized(w http.ResponseWriter) {
	metadata := h.endpoint.Scheme + "://" + h.endpoint.Host + MetadataPath
	w.Header().Set("WWW-Authenticate", `Bearer resource_metadata="`+metadata+`"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
func (h *handler) server(requestContext context.Context, principal identity.Principal) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "Stuff Stash", Version: "1"}, &mcp.ServerOptions{SupportedProtocolVersions: []string{"2026-07-28", "2025-11-25"}})
	for _, definition := range tools.ReadCatalog() {
		mcp.AddTool(server, &mcp.Tool{Name: string(definition.Name), Description: definition.Description, InputSchema: definition.InputSchema, OutputSchema: definition.OutputSchema, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, req *mcp.CallToolRequest, input arguments) (*mcp.CallToolResult, map[string]any, error) {
			// The SDK detaches HTTP cancellation for older protocol versions. This
			// stateless server must retain both the carrier deadline and tool cancellation.
			callContext, cancel := context.WithCancel(requestContext)
			stop := context.AfterFunc(ctx, cancel)
			defer stop()
			defer cancel()
			output, err := h.call(callContext, principal, definition, input)
			return nil, output, err
		})
	}

	return server
}
