package bootstrap

import (
	"testing"

	"github.com/stuffstash/stuff-stash/internal/config"
)

func TestValidateInvitationPublicBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Config
		wantErr bool
	}{
		{name: "local empty", config: config.Config{AuthMode: "local-dev"}},
		{name: "local HTTP requires explicit switch", config: config.Config{AuthMode: "local-dev", InvitationPublicBaseURL: "http://localhost:5173/invitations/accept"}, wantErr: true},
		{name: "explicit local loopback", config: config.Config{AuthMode: "local-dev", InvitationPublicBaseURL: "http://127.0.0.1:5173/invitations/accept", InvitationAllowInsecureLocalHTTP: true}},
		{name: "explicit local oidc loopback", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "http://localhost:5173/invitations/accept", InvitationAllowInsecureLocalHTTP: true}},
		{name: "explicit private LAN 10", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "http://10.0.0.7:5173/invitations/accept", InvitationAllowInsecureLocalHTTP: true}},
		{name: "explicit private LAN 172", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "http://172.31.255.254:5173/invitations/accept", InvitationAllowInsecureLocalHTTP: true}},
		{name: "explicit private LAN 192", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "http://192.168.1.117:5173/invitations/accept", InvitationAllowInsecureLocalHTTP: true}},
		{name: "private LAN requires explicit switch", config: config.Config{AuthMode: "local-dev", InvitationPublicBaseURL: "http://192.168.1.117:5173/invitations/accept"}, wantErr: true},
		{name: "public IP remains forbidden", config: config.Config{AuthMode: "local-dev", InvitationPublicBaseURL: "http://8.8.8.8:5173/invitations/accept", InvitationAllowInsecureLocalHTTP: true}, wantErr: true},
		{name: "outside private 172 range remains forbidden", config: config.Config{AuthMode: "local-dev", InvitationPublicBaseURL: "http://172.32.0.1:5173/invitations/accept", InvitationAllowInsecureLocalHTTP: true}, wantErr: true},
		{name: "production https", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "https://stash.example.test/invitations/accept"}},
		{name: "production omitted", config: config.Config{AuthMode: "oidc"}, wantErr: true},
		{name: "production localhost", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "http://localhost:5173/invitations/accept"}, wantErr: true},
		{name: "production remote http", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "http://stash.example.test/invitations/accept"}, wantErr: true},
		{name: "userinfo", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "https://user:pass@stash.example.test/invitations/accept"}, wantErr: true},
		{name: "query", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "https://stash.example.test/invitations/accept?unsafe=yes"}, wantErr: true},
		{name: "fragment", config: config.Config{AuthMode: "oidc", InvitationPublicBaseURL: "https://stash.example.test/invitations/accept#unsafe"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInvitationPublicBaseURL(tc.config)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validate invitation public base URL: err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
