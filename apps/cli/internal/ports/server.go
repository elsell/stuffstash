package ports

import "context"

type ServerInfo struct {
	InstanceID      string `json:"instanceId"`
	ProtocolVersion int64  `json:"protocolVersion"`
}
type ServerAuthConfig struct {
	Issuer           string           `json:"issuer"`
	ClientID         string           `json:"clientId"`
	Scopes           []string         `json:"scopes"`
	LoginMethods     []string         `json:"loginMethods"`
	LoopbackRedirect LoopbackRedirect `json:"loopbackRedirect"`
}
type LoopbackRedirect struct {
	Host          string `json:"host"`
	PathPrefix    string `json:"pathPrefix"`
	EphemeralPort bool   `json:"ephemeralPort"`
}
type ServerAPI interface {
	ServerInfo(context.Context) (Result[ServerInfo], error)
	ServerAuthConfig(context.Context) (Result[ServerAuthConfig], error)
}
