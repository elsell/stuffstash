package presentation

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (o Output) serverAuthConfig(v ports.ServerAuthConfig) error {
	fields := [][2]string{{"Issuer", v.Issuer}, {"Client ID", v.ClientID}, {"Loopback host", v.LoopbackRedirect.Host}, {"Callback path prefix", v.LoopbackRedirect.PathPrefix}, {"Ephemeral port", strconv.FormatBool(v.LoopbackRedirect.EphemeralPort)}}
	for _, scope := range v.Scopes {
		fields = append(fields, [2]string{"Scope", scope})
	}
	for _, method := range v.LoginMethods {
		fields = append(fields, [2]string{"Login method", method})
	}
	return o.details(fields)
}
