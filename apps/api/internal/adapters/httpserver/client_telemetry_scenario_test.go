package httpserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func coverClientTelemetryScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	observer := &fakeObserver{}
	server := NewServer(":0", newTestApp(observer))
	authorization, status := "Bearer dev:owner", http.StatusOK
	if adversarial {
		authorization = "Bearer private-malformed"
		status = http.StatusUnauthorized
	}
	response := coverage.request(t, server, http.MethodPost, "/client-telemetry", "/client-telemetry", authorization, map[string]any{"measurements": []any{
		map[string]any{"platform": "web", "operation": "image", "surface": "gallery", "variant": "medium", "outcome": "success", "durationMs": 125.5},
	}}, status)
	if adversarial {
		if len(observer.events) != 0 {
			t.Fatal("unauthorized telemetry emitted events")
		}
		return
	}
	var result struct {
		Data struct {
			Accepted int `json:"accepted"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Data.Accepted != 1 {
		t.Fatal("accepted measurement count missing")
	}
}
