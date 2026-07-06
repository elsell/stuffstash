package httpserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestMobileAuthMetadataReturnsPublicOIDCConfiguration(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		MobileAuth: MobileAuthOptions{
			Issuer:      "https://accounts.example.test/",
			ClientID:    "stuff-stash-mobile",
			RedirectURI: "stuffstash://auth/callback",
			Scopes:      []string{"openid", "email", "profile", "offline_access", "email"},
		},
	})

	response := performRequest(server, http.MethodGet, "/.well-known/stuff-stash/mobile-auth", "", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}
	rawBody := response.Body.String()
	if rawBody == "" || strings.Contains(rawBody, "secret") || strings.Contains(rawBody, "password") {
		t.Fatalf("mobile auth metadata leaked unsafe material: %s", rawBody)
	}

	var body struct {
		Data mobileAuthMetadataResponse `json:"data"`
		Meta responseMeta               `json:"meta"`
	}
	decodeBody(t, response, &body)

	if body.Data.Issuer != "https://accounts.example.test" {
		t.Fatalf("expected normalized issuer, got %q", body.Data.Issuer)
	}
	if body.Data.ClientID != "stuff-stash-mobile" || body.Data.RedirectURI != "stuffstash://auth/callback" {
		t.Fatalf("unexpected mobile auth metadata: %+v", body.Data)
	}
	if len(body.Data.Scopes) != 4 || body.Data.Scopes[0] != "openid" || body.Data.Scopes[3] != "offline_access" {
		t.Fatalf("unexpected mobile scopes: %+v", body.Data.Scopes)
	}
}

func TestMobileAuthMetadataFailsClosedWhenUnavailable(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequest(server, http.MethodGet, "/.well-known/stuff-stash/mobile-auth", "", nil)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusServiceUnavailable, response.Code, response.Body.String())
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	decodeBody(t, response, &body)

	if body.Error.Code != "mobile_auth_unavailable" || body.Error.Message != "Mobile sign-in is not configured." {
		t.Fatalf("unexpected mobile auth error: %+v", body.Error)
	}
}

func TestMobileAuthMetadataFailsClosedForUnsupportedRedirectURI(t *testing.T) {
	server := NewServerWithOptions(":0", newTestApp(&fakeObserver{}, "unused-id"), Options{
		MobileAuth: MobileAuthOptions{
			Issuer:      "https://accounts.example.test",
			ClientID:    "stuff-stash-mobile",
			RedirectURI: "https://evil.example.test/callback",
			Scopes:      []string{"openid"},
		},
	})

	response := performRequest(server, http.MethodGet, "/.well-known/stuff-stash/mobile-auth", "", nil)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusServiceUnavailable, response.Code, response.Body.String())
	}
}
