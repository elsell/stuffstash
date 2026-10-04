package voice

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLocalAcceptanceEvidencePreservesResponseAndExcludesContent(t *testing.T) {
	for _, finish := range []string{"length", "tool_calls", "private-model-text"} {
		t.Run(finish, func(t *testing.T) {
			raw := `{"choices":[{"finish_reason":"` + finish + `","message":{"content":"private-content","tool_calls":[{"function":{"arguments":"private-arguments"}}]}}]}`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, raw)
			}))
			defer server.Close()
			evidence := &localResponseEvidence{}
			client := &http.Client{Transport: localEvidenceTransport{evidence}}
			response, err := client.Get(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			got, err := io.ReadAll(response.Body)
			if err != nil || string(got) != raw {
				t.Fatal("observer changed provider response")
			}
			response.Body.Close()
			encoded, err := json.Marshal(evidence)
			if err != nil || strings.Contains(string(encoded), "private-") {
				t.Fatal("response content leaked into evidence")
			}
			expectedFinish := finish
			if finish == "private-model-text" {
				expectedFinish = "invalid-or-unknown"
			}
			if evidence.Status != 200 || evidence.ToolCalls != 1 || evidence.Finish != expectedFinish {
				t.Fatalf("incorrect structural evidence: %s", encoded)
			}
		})
	}
}

func TestLocalAcceptanceEvidenceDistinguishesDeadline(t *testing.T) {
	if localAcceptanceError(context.DeadlineExceeded) != "deadline-exceeded" {
		t.Fatal("deadline not classified")
	}
}

func TestLocalEvidenceDoesNotReadRejectedResponseBody(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.(http.Flusher).Flush()
		<-release
	}))
	defer server.Close()
	defer close(release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (localEvidenceTransport{&localResponseEvidence{}}).RoundTrip(req)
	if err != nil {
		t.Fatalf("observer waited for rejected body: %v", err)
	}
	response.Body.Close()
}

func TestLocalEvidencePreservesInterruptedRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = io.WriteString(w, "short")
	}))
	defer server.Close()
	client := &http.Client{Transport: localEvidenceTransport{&localResponseEvidence{}}}
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("observer moved body failure into request: %v", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if string(raw) != "short" || err != io.ErrUnexpectedEOF {
		t.Fatalf("read changed: %q %v", raw, err)
	}
}
