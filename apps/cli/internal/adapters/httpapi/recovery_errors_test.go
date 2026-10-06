package httpapi

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIRecoveryPreservesCategoriesAndHidesResponseBodies(t *testing.T) {
	cases := []struct {
		status           int
		category, advice string
	}{
		{400, "validation", "Check the input"}, {422, "validation", "Check the input"},
		{401, "authentication", "stuffstash login"}, {403, "forbidden", "owner"},
		{404, "not_found", "household and inventory"}, {409, "conflict", "current state"},
		{429, "api", "Wait"}, {503, "unavailable", "before you repeat"}, {500, "api", "before you repeat"},
		{200, "protocol", "before you repeat"},
	}
	for _, test := range cases {
		t.Run(http.StatusText(test.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				io.WriteString(w, "private-token-and-body")
			}))
			defer server.Close()
			api, _ := New(server.URL, "token", server.Client())
			_, err := api.ServerInfo(context.Background())
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != test.category || !strings.Contains(failure.Message, test.advice) || strings.Contains(failure.Message, "private-token-and-body") {
				t.Fatalf("unsafe or unhelpful error: %v", err)
			}
		})
	}
	server := httptest.NewServer(http.NotFoundHandler())
	client := server.Client()
	server.Close()
	api, _ := New(server.URL, "token", client)
	_, err := api.ServerInfo(context.Background())
	var failure *ports.Error
	if !errors.As(err, &failure) || failure.Category != "network" || !strings.Contains(failure.Message, "connection") || !strings.Contains(failure.Message, "before you repeat") || strings.Contains(failure.Message, server.URL) {
		t.Fatalf("network recovery: %v", err)
	}
}
