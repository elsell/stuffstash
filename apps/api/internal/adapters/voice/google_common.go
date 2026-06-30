package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const googleCloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
const googleDefaultHTTPTimeout = 60 * time.Second

type googleHTTPClient struct {
	baseURL      string
	httpClient   *http.Client
	tokenSource  oauth2.TokenSource
	quotaProject string
	apiKey       string
}

func newGoogleHTTPClient(baseURL string, httpClient *http.Client, httpTimeout time.Duration, tokenSource oauth2.TokenSource, quotaProject string, apiKey string) googleHTTPClient {
	if httpClient == nil {
		if httpTimeout <= 0 {
			httpTimeout = googleDefaultHTTPTimeout
		}
		httpClient = &http.Client{Timeout: httpTimeout}
	}
	return googleHTTPClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		httpClient:   httpClient,
		tokenSource:  tokenSource,
		quotaProject: strings.TrimSpace(quotaProject),
		apiKey:       strings.TrimSpace(apiKey),
	}
}

func (c googleHTTPClient) postJSON(ctx context.Context, path string, request any, response any) error {
	if c.tokenSource == nil && c.apiKey == "" {
		return errors.New("google token source is not configured")
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if c.apiKey != "" {
		httpRequest.Header.Set("X-Goog-Api-Key", c.apiKey)
	} else {
		token, err := c.tokenSource.Token()
		if err != nil {
			return err
		}
		httpRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if c.apiKey == "" && c.quotaProject != "" {
		httpRequest.Header.Set("X-Goog-User-Project", c.quotaProject)
	}
	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return googleProviderTimeoutError{}
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return googleProviderTimeoutError{}
		}
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return googleProviderHTTPError{statusCode: httpResponse.StatusCode}
	}
	return json.NewDecoder(httpResponse.Body).Decode(response)
}

type googleProviderTimeoutError struct{}

func (googleProviderTimeoutError) Error() string {
	return "google provider request timed out"
}

func (googleProviderTimeoutError) SafeRealtimeVoiceDiagnostic() string {
	return "provider_timeout"
}

type googleProviderHTTPError struct {
	statusCode int
}

func (e googleProviderHTTPError) Error() string {
	return fmt.Sprintf("google provider request failed with status %d", e.statusCode)
}

func (e googleProviderHTTPError) SafeRealtimeVoiceDiagnostic() string {
	return fmt.Sprintf("provider_http_status_%d", e.statusCode)
}

func firstGeminiText(response geminiGenerateContentResponse) string {
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			text := strings.TrimSpace(part.Text)
			if text != "" {
				return text
			}
		}
	}
	return ""
}

func geminiFunctionCalls(response geminiGenerateContentResponse) []geminiFunctionCall {
	calls := []geminiFunctionCall{}
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.FunctionCall != nil {
				calls = append(calls, *part.FunctionCall)
			}
		}
	}
	return calls
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}
