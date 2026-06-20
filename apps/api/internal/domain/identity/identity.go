package identity

import (
	"strings"
	"unicode"
)

type PrincipalID string

func NewPrincipalID(value string) (PrincipalID, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", false
	}

	return PrincipalID(value), true
}

func (id PrincipalID) String() string {
	return string(id)
}

type Principal struct {
	ID    PrincipalID
	Email Email
}

type Email string

func NewEmail(value string) (Email, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || strings.ContainsAny(value, " \t\r\n") {
		return "", false
	}
	parts := strings.Split(value, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return "", false
	}
	return Email(value), true
}

func (e Email) String() string {
	return string(e)
}
