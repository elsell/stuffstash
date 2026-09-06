package httpserver

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type voiceTraceRecorder struct{ lines []string }

func (r *voiceTraceRecorder) Logf(format string, args ...any) {
	r.lines = append(r.lines, fmt.Sprintf(format, args...))
}
func TestNativeVoiceTraceRetainsToolEvidenceWithoutThoughtsOrCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"candidates":[{"content":{"parts":[{"thought":true,"text":"private-reasoning","thoughtSignature":"private-signature"},{"functionCall":{"name":"search","args":{"query":"baby clothes"}}},{"text":"A useful answer."}]}}]}`)
	}))
	defer server.Close()
	body := `{"contents":[{"parts":[{"functionResponse":{"name":"search","response":{"output":{"count":2}}}},{"text":"{\"toolResults\":[{\"callId\":\"search-2\",\"name\":\"search\",\"output\":{\"count\":3}}]}"},{"thought":true,"text":"{\"toolResults\":[\"private-reasoning\"]}"},{"text":"private-prompt"}]}]}`
	request, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer private-token")
	recorder := &voiceTraceRecorder{}
	transport := liveVoiceTraceTransport{t: recorder, next: http.DefaultTransport}
	response, err := transport.RoundTrip(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	restored, err := io.ReadAll(response.Body)
	if err != nil || !strings.Contains(string(restored), "private-signature") {
		t.Fatal("tracing damaged the provider response")
	}
	logs := strings.Join(recorder.lines, "\n")
	for _, secret := range []string{"private-reasoning", "private-signature", "private-token", "private-prompt"} {
		if strings.Contains(logs, secret) {
			t.Fatalf("private provider material logged: %s", secret)
		}
	}
	for _, evidence := range []string{"baby clothes", `"count":2`, `"count":3`, "search-2", "A useful answer.", "requestBytes", "responseBytes"} {
		if !strings.Contains(logs, evidence) {
			t.Fatalf("missing trace evidence %q", evidence)
		}
	}
}
