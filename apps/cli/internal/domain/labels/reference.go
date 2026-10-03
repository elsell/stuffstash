package labels

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

type Reference struct{ Instance, Label string }

var pathPattern = regexp.MustCompile(`^(?:/[A-Za-z0-9._~-]+)*/l/(v[0-9]+)/([^/]+)/([^/]+)$`)
var appPattern = regexp.MustCompile(`^/(v[0-9]+)/([^/]+)/([^/]+)$`)
var opaqueID = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)
var dots = regexp.MustCompile(`(?:^|/)\.{1,2}(?:/|$)`)

func Parse(value string) (Reference, error) {
	invalid := errors.New("invalid label link")
	if len(value) > 4096 || strings.ContainsAny(value, `\%?#`) || strings.IndexFunc(value, unicode.IsSpace) >= 0 || dots.MatchString(value) {
		return Reference{}, invalid
	}
	u, err := url.Parse(value)
	if err != nil || u.User != nil || u.Opaque != "" {
		return Reference{}, invalid
	}
	var match []string
	switch {
	case u.Scheme == "https" && u.Hostname() != "" && !strings.Contains(u.Path, "//"):
		match = pathPattern.FindStringSubmatch(u.Path)
	case u.Scheme == "stuffstash" && u.Host == "labels":
		match = appPattern.FindStringSubmatch(u.Path)
	}
	if len(match) != 4 || match[1] != "v1" || !opaqueID.MatchString(match[2]) || !opaqueID.MatchString(match[3]) {
		return Reference{}, invalid
	}
	return Reference{Instance: match[2], Label: match[3]}, nil
}
