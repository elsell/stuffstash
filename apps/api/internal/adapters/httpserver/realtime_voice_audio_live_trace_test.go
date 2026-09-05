package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// Records only bounded model text for the controlled fixture test. Credentials,
// request headers, token exchanges and provider error bodies are never logged.
type liveVoiceTraceTransport struct {
	t    *testing.T
	next http.RoundTripper
}

func (transport liveVoiceTraceTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := transport.next.RoundTrip(request)
	if err != nil || response.StatusCode != http.StatusOK {
		return response, err
	}
	const maxCapturedBytes = 256 * 1024
	body, err := io.ReadAll(io.LimitReader(response.Body, maxCapturedBytes+1))
	response.Body = &liveVoiceRestoredBody{Reader: io.MultiReader(bytes.NewReader(body), response.Body), Closer: response.Body}
	if err != nil || len(body) > maxCapturedBytes {
		return response, nil
	}
	var payload struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if json.Unmarshal(body, &payload) == nil {
		for _, candidate := range payload.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					transport.t.Logf("VOICE_PROVIDER_TEXT %s", part.Text)
				}
			}
		}
	}
	return response, nil
}

type liveVoiceRestoredBody struct {
	io.Reader
	io.Closer
}
