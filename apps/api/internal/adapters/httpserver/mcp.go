package httpserver

import "net/http"

// MCP owns its protocol/metadata responses; the outer server owns shared middleware.
func registerMCPHandler(mux *http.ServeMux, handler http.Handler) {
	if handler == nil {
		return
	}
	mux.Handle("/mcp", handler)
	mux.Handle("/.well-known/oauth-protected-resource/mcp", handler)
}
