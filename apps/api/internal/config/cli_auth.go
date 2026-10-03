package config

func oidcCLIScopes() []string {
	scopes := stringListEnv(envOIDCCLIScopes)
	if len(scopes) == 0 {
		return []string{"openid", "email", "profile", "offline_access"}
	}
	return scopes
}
