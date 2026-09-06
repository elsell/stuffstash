package httpserver

import (
	"net/http"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestClientTelemetryAuthenticationAndValidation(t *testing.T) {
	valid := map[string]any{"platform": "web", "operation": "image", "surface": "gallery", "variant": "medium", "outcome": "success", "durationMs": 125.5}
	oversized := make([]any, 51)
	for i := range oversized {
		oversized[i] = valid
	}
	for _, test := range []struct {
		name, authorization string
		measurements        any
		want                int
	}{
		{"application transport", "Bearer dev:owner", []any{map[string]any{"platform": "web", "operation": "request", "surface": "application", "variant": "none", "outcome": "success", "durationMs": 12}}, 200},
		{"authenticated", "Bearer dev:owner", []any{valid}, 200},
		{"other authenticated principal", "Bearer dev:other", []any{valid}, 200},
		{"oversized", "Bearer dev:owner", oversized, 422},
		{"mixed invalid", "Bearer dev:owner", []any{valid, map[string]any{"platform": "private"}}, 422},
		{"anonymous", "", []any{valid}, 401},
		{"malformed token", "Bearer private-invalid-token", []any{valid}, 401},
		{"empty", "Bearer dev:owner", []any{}, 422},
		{"unbounded dimension", "Bearer dev:owner", []any{map[string]any{"platform": "private-tenant", "operation": "image", "surface": "gallery", "variant": "medium", "outcome": "success", "durationMs": 1}}, 422},
		{"negative duration", "Bearer dev:owner", []any{map[string]any{"platform": "web", "operation": "image", "surface": "gallery", "variant": "medium", "outcome": "success", "durationMs": -1}}, 422},
	} {
		t.Run(test.name, func(t *testing.T) {
			observer := &fakeObserver{}
			server := NewServer(":0", newTestApp(observer))
			response := performRequest(server, http.MethodPost, "/client-telemetry", test.authorization, map[string]any{"measurements": test.measurements})
			if response.Code != test.want {
				t.Fatalf("status %d; expected %d", response.Code, test.want)
			}
			count := 0
			for _, event := range observer.events {
				if event.Name == ports.EventName("client.performance.observed") {
					count++
					if len(event.Fields) != 6 {
						t.Fatal("unexpected client telemetry fields")
					}
				}
			}
			if test.want == 200 && count != 1 {
				t.Fatal("accepted measurement not observed")
			}
			if test.want != 200 && count != 0 {
				t.Fatal("rejected measurement observed")
			}
		})
	}
}
