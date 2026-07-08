package voice

import (
	"regexp"
	"strings"
)

func safeGoogleConversationPromptText(value string, maxLength int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = googleConversationURLPattern.ReplaceAllString(value, "[redacted-url]")
	value = googleConversationRawResponseTailPattern.ReplaceAllString(value, "[redacted]")
	value = googleConversationAssignmentPattern.ReplaceAllString(value, "[redacted]")
	value = googleConversationBearerPattern.ReplaceAllString(value, "bearer [redacted]")
	replacer := strings.NewReplacer(
		"raw prompt", "[redacted]",
		"Raw prompt", "[redacted]",
		"raw provider response", "[redacted]",
		"Raw provider response", "[redacted]",
		"raw model response", "[redacted]",
		"Raw model response", "[redacted]",
		"bearer", "[redacted]",
		"Bearer", "[redacted]",
		"credential", "[redacted]",
		"Credential", "[redacted]",
		"password", "[redacted]",
		"Password", "[redacted]",
		"secret", "[redacted]",
		"Secret", "[redacted]",
		"token", "[redacted]",
		"Token", "[redacted]",
		"api key", "[redacted]",
		"API key", "[redacted]",
		"apikey", "[redacted]",
	)
	value = replacer.Replace(value)
	if len(value) <= maxLength {
		return value
	}
	return strings.TrimSpace(value[:maxLength]) + " ..."
}

var googleConversationAssignmentPattern = regexp.MustCompile(`(?i)\b(api[-_ ]?key|authorization|credential|password|provider[-_ ]?session[-_ ]?id|providerSessionId|secret|token)\s*[:=]\s*["']?[^"',\s}\n]+`)
var googleConversationBearerPattern = regexp.MustCompile(`(?i)\bbearer\s+[a-z0-9._~+/=-]+`)
var googleConversationRawResponseTailPattern = regexp.MustCompile(`(?is)\b(raw[-_ ]?(model[-_ ]?response|provider[-_ ]?response)|raw\s+(model|provider)\s+response)\b.*`)
var googleConversationURLPattern = regexp.MustCompile(`(?i)\b(?:https?|wss?)://[^\s"',\]}]+`)
