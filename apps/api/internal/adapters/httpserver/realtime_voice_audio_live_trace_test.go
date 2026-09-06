package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Records bounded native tool evidence for controlled fixture tests. Credentials,
// thoughts, signatures, headers, audio and provider error bodies are never logged.
type liveVoiceTraceTransport struct {
	t    interface{ Logf(string, ...any) }
	next http.RoundTripper
}

func (transport liveVoiceTraceTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	const maxCapturedBytes = 256 * 1024
	if request.GetBody != nil {
		copy, err := request.GetBody()
		if err == nil {
			body, err := io.ReadAll(io.LimitReader(copy, maxCapturedBytes+1))
			_ = copy.Close()
			if err == nil && len(body) <= maxCapturedBytes {
				transport.logNativeParts(body)
			}
		}
	}
	started := time.Now()
	response, err := transport.next.RoundTrip(request)
	if err != nil || response.StatusCode != http.StatusOK {
		return response, err
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxCapturedBytes+1))
	response.Body = &liveVoiceRestoredBody{Reader: io.MultiReader(bytes.NewReader(body), response.Body), Closer: response.Body}
	if err != nil || len(body) > maxCapturedBytes {
		return response, nil
	}
	transport.t.Logf("VOICE_PROVIDER_IO requestBytes=%d responseBytes=%d elapsedMs=%d", request.ContentLength, len(body), time.Since(started).Milliseconds())
	transport.logNativeParts(body)
	return response, nil
}

type liveVoiceRestoredBody struct {
	io.Reader
	io.Closer
}

type liveVoiceNativeContent struct {
	Parts []struct {
		Text             string          `json:"text"`
		Thought          bool            `json:"thought"`
		FunctionCall     json.RawMessage `json:"functionCall"`
		FunctionResponse json.RawMessage `json:"functionResponse"`
	} `json:"parts"`
}

func (transport liveVoiceTraceTransport) logNativeParts(body []byte) {
	var payload struct {
		Contents   []liveVoiceNativeContent `json:"contents"`
		Candidates []struct {
			Content liveVoiceNativeContent `json:"content"`
		} `json:"candidates"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return
	}
	// Outgoing requests expose only tool responses, never prompts/audio.
	for _, content := range payload.Contents {
		for _, part := range content.Parts {
			if part.Thought {
				continue
			}
			if len(part.FunctionResponse) > 0 {
				transport.t.Logf("VOICE_TOOL_RESULT %s", part.FunctionResponse)
			}
			var envelope struct {
				ToolResults []json.RawMessage `json:"toolResults"`
			}
			if part.Text != "" && json.Unmarshal([]byte(part.Text), &envelope) == nil {
				for _, result := range envelope.ToolResults {
					transport.t.Logf("VOICE_TOOL_RESULT %s", result)
				}
			}
		}
	}
	for _, candidate := range payload.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Thought {
				continue
			}
			if part.Text != "" {
				transport.t.Logf("VOICE_PROVIDER_TEXT %s", part.Text)
			}
			if len(part.FunctionCall) > 0 {
				transport.t.Logf("VOICE_TOOL_CALL %s", part.FunctionCall)
			}
		}
	}
}
