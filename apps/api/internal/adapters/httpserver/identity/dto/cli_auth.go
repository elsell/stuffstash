package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

// CLIAuthMetadata contains only public settings for a separately registered OIDC client.
type CLIAuthMetadata struct {
	Issuer           string              `json:"issuer"`
	ClientID         string              `json:"clientId"`
	Scopes           []string            `json:"scopes"`
	LoginMethods     []string            `json:"loginMethods"`
	LoopbackRedirect CLILoopbackRedirect `json:"loopbackRedirect"`
}

type CLILoopbackRedirect struct {
	Host          string `json:"host"`
	PathPrefix    string `json:"pathPrefix"`
	EphemeralPort bool   `json:"ephemeralPort"`
}

type CLIAuthConfigInput struct{}

type CLIAuthConfigOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[CLIAuthMetadata]
}
