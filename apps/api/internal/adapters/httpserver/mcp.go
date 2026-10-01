package httpserver

import "net/http"

// MCP owns its protocol/metadata responses; the outer server owns shared middleware.
func registerMCPHandler(mux *http.ServeMux, handler http.Handler) {
	if handler == nil {
		return
	}
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		mux.Handle(method+" /mcp", handler)
	}
	mux.Handle("GET /.well-known/oauth-protected-resource/mcp", handler)
}
