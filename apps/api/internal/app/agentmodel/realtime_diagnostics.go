package agentmodel

import (
	"regexp"
	"strings"
)

func EmitRealtimeVoiceDiagnostic(sessionID string, title string, detail string, emit RealtimeVoiceEventSink) error {
	message := SafeRealtimeVoiceDiagnosticText(title, 120)
	if message == "" {
		message = "Agent diagnostic"
	}
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAgentDiagnostic, SessionID: sessionID, Message: message, Detail: SafeRealtimeVoiceDiagnosticText(detail, 4000)})
}

func SafeRealtimeVoiceDiagnosticText(value string, maxLength int) string {
	trimmed := strings.TrimSpace(redactRealtimeVoiceDiagnosticString(value))
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= maxLength {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:maxLength]) + " ..."
}

func redactRealtimeVoiceDiagnosticString(value string) string {
	value = realtimeVoiceDiagnosticURLPattern.ReplaceAllString(value, "[redacted-url]")
	value = realtimeVoiceDiagnosticBearerPattern.ReplaceAllString(value, "[redacted-bearer] [redacted]")
	value = realtimeVoiceDiagnosticAssignmentPattern.ReplaceAllString(value, "$1[redacted]")
	value = realtimeVoiceDiagnosticRawResponseAssignmentPattern.ReplaceAllString(value, "[redacted]")
	value = realtimeVoiceDiagnosticUnsafePhrasePattern.ReplaceAllString(value, "[redacted]")
	replacer := strings.NewReplacer(
		"apiKey", "[redacted-key]",
		"api_key", "[redacted-key]",
		"authorization", "[redacted-authorization]",
		"credential", "[redacted-credential]",
		"password", "[redacted-password]",
		"providerSessionId", "[redacted-provider-session]",
		"secret", "[redacted-secret]",
		"token", "[redacted-token]",
	)
	return replacer.Replace(value)
}

var realtimeVoiceDiagnosticAssignmentPattern = regexp.MustCompile(`(?i)\b(api[-_ ]?key|authorization|credential|password|provider[-_ ]?session[-_ ]?id|secret|token)\s*[:=]\s*["']?[^"',\s}\n]+`)
var realtimeVoiceDiagnosticBearerPattern = regexp.MustCompile(`(?i)\b(bearer)\s+[a-z0-9._~+/=-]+`)
var realtimeVoiceDiagnosticRawResponseAssignmentPattern = regexp.MustCompile(`(?i)\b(raw[-_ ]?(model[-_ ]?response|provider[-_ ]?response)|raw\s+(model|provider)\s+response)\s*[:=]\s*[^;\n\r]+`)
var realtimeVoiceDiagnosticUnsafePhrasePattern = regexp.MustCompile(`(?i)\b(raw[-_ ]?(prompt|query|transcript|model[-_ ]?response|provider[-_ ]?response)|stack[-_ ]?trace|provider[-_ ]+session[-_ ]+id)\b`)
var realtimeVoiceDiagnosticURLPattern = regexp.MustCompile(`(?i)\b(?:https?|wss?)://[^\s"',\]}]+`)
