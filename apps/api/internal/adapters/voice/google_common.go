package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const googleCloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

type googleHTTPClient struct {
	baseURL      string
	httpClient   *http.Client
	tokenSource  oauth2.TokenSource
	quotaProject string
}

func newGoogleHTTPClient(baseURL string, httpClient *http.Client, tokenSource oauth2.TokenSource, quotaProject string) googleHTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return googleHTTPClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		httpClient:   httpClient,
		tokenSource:  tokenSource,
		quotaProject: strings.TrimSpace(quotaProject),
	}
}

func (c googleHTTPClient) postJSON(ctx context.Context, path string, request any, response any) error {
	if c.tokenSource == nil {
		return errors.New("google token source is not configured")
	}
	token, err := c.tokenSource.Token()
	if err != nil {
		return err
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	httpRequest.Header.Set("Content-Type", "application/json")
	if c.quotaProject != "" {
		httpRequest.Header.Set("X-Goog-User-Project", c.quotaProject)
	}
	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return errors.New("google provider request failed")
	}
	return json.NewDecoder(httpResponse.Body).Decode(response)
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

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}
