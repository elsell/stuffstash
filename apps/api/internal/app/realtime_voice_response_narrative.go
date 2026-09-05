package app

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

var realtimeVoiceNarrativeForbiddenPhrases = [...]string{"visible match", "candidate", "resolution", "tool result", "tool call", "asset id", "inventory id", "tenant id"}

// Remove complete grounded labels only for presentation-style checks. All
// factual, provenance and final-response safety checks see the original text.
func realtimeVoiceResponseNarrative(brief agentmodel.GroundedVoiceResponseBrief, text string) string {
	labels := []string{}
	for _, finding := range brief.Findings {
		labels = append(labels, finding.Title)
		labels = append(labels, finding.ContainmentPath...)
	}
	for i := range labels {
		labels[i] = strings.ToLower(strings.TrimSpace(labels[i]))
	}
	sort.Slice(labels, func(i, j int) bool { return len(labels[i]) > len(labels[j]) })
	var result strings.Builder
	for offset := 0; offset < len(text); {
		matched := ""
		for _, label := range labels {
			// Punctuation alone cannot distinguish an entity from sentence syntax.
			if strings.IndexFunc(label, unicode.IsLetter) < 0 && strings.IndexFunc(label, unicode.IsNumber) < 0 {
				continue
			}
			if !strings.HasPrefix(text[offset:], label) {
				continue
			}
			before, _ := utf8.DecodeLastRuneInString(text[:offset])
			after, _ := utf8.DecodeRuneInString(text[offset+len(label):])
			if unicode.IsLetter(before) || unicode.IsNumber(before) || unicode.IsLetter(after) || unicode.IsNumber(after) {
				continue
			}
			if realtimeVoiceLabelSplitsDiagnostic(text, offset, offset+len(label)) {
				continue
			}
			matched = label
			break
		}
		if matched != "" {
			result.WriteString("[asset]")
			offset += len(matched)
			continue
		}
		r, size := utf8.DecodeRuneInString(text[offset:])
		result.WriteRune(r)
		offset += size
	}
	return result.String()
}

// A short title cannot hide part of a multiword diagnostic phrase.
func realtimeVoiceLabelSplitsDiagnostic(text string, start, end int) bool {
	for _, phrase := range realtimeVoiceNarrativeForbiddenPhrases {
		for offset := 0; offset < len(text); {
			index := strings.Index(text[offset:], phrase)
			if index < 0 {
				break
			}
			left := offset + index
			right := left + len(phrase)
			if left < end && right > start && (left < start || right > end) {
				return true
			}
			offset = right
		}
	}
	return false
}
