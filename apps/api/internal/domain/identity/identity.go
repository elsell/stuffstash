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
	ID PrincipalID
}
