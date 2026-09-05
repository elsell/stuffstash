package agentmodel

import (
	"strings"
	"unicode/utf8"
)

const MaxObservationTagNames = 32
const MaxObservationTagNameBytes = 80

// BoundedObservationTagNames preserves stable, distinct inventory labels within
// the evidence budget. Labels are data, not model instructions.
func BoundedObservationTagNames(names []string) []string {
	result := make([]string, 0, min(len(names), MaxObservationTagNames))
	seen := map[string]struct{}{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || len(name) > MaxObservationTagNameBytes || !utf8.ValidString(name) {
			continue
		}
		key := strings.ToLower(name)
		if _, found := seen[key]; found {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, name)
		if len(result) == MaxObservationTagNames {
			break
		}
	}
	return result
}
func validObservationTagNames(names []string) bool {
	if len(names) > MaxObservationTagNames {
		return false
	}
	normalized := BoundedObservationTagNames(names)
	if len(normalized) != len(names) {
		return false
	}
	for i := range names {
		if names[i] != normalized[i] {
			return false
		}
	}
	return true
}
