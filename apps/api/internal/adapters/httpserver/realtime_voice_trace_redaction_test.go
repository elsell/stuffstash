package httpserver

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type capturedLiveVoiceLogs struct{ lines []string }

func (l *capturedLiveVoiceLogs) Logf(format string, args ...any) {
	l.lines = append(l.lines, fmt.Sprintf(format, args...))
}
func TestLiveVoiceSchemaErrorClassificationDoesNotExposeBody(t *testing.T) {
	body := `{"error":{"message":"schema too complex; private-fixture-secret"}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(400); _, _ = io.WriteString(w, body) }))
	defer server.Close()
	logs := &capturedLiveVoiceLogs{}
	transport := liveVoiceTraceTransport{t: logs, next: http.DefaultTransport}
	request, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	response, err := transport.RoundTrip(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	restored, err := io.ReadAll(response.Body)
	if err != nil || string(restored) != body {
		t.Fatal("diagnostic consumed or changed response body")
	}
	joined := strings.Join(logs.lines, "\n")
	if !strings.Contains(joined, "categories=[schema complex]") || strings.Contains(joined, "private-fixture-secret") || strings.Contains(joined, body) {
		t.Fatalf("unsafe diagnostic classification: %s", joined)
	}
}
