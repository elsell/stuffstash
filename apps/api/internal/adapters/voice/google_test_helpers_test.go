package voice

import (
	"testing"

	"golang.org/x/oauth2"
)

func objectAt(t *testing.T, item map[string]any, key string) map[string]any {
	t.Helper()
	return objectFromAny(t, item[key])
}

func objectFromAny(t *testing.T, value any) map[string]any {
	t.Helper()
	item, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is not an object: %+v", value)
	}
	return item
}

type staticTokenSource struct{}

func (staticTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "test-token", TokenType: "Bearer"}, nil
}
