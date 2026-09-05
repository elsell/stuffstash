package agentmodel

import (
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"strings"
)

// RenderGroundedVoiceResponse renders only facts already authorized and resolved
// by the application. It neither invokes a model nor performs inventory actions.
func RenderGroundedVoiceResponse(brief domain.GroundedVoiceResponseBrief, maxBytes int) (string, error) {
	if brief.Validate() != nil || maxBytes < 160 {
		return "", domain.ErrInvalidGroundedVoiceResponseBrief
	}
	var rows []string
	prefix := ""
	switch brief.Mode {
	case domain.ResponseAnswerModeNotFound:
		rows = []string{"I couldn't find " + brief.Subject + " in the searched inventory results."}
	case domain.ResponseAnswerModeUnsupported:
		rows = []string{"I can't perform that inventory action. You can ask me to find an item or describe an inventory change."}
	case domain.ResponseAnswerModeClarify:
		question := "Which " + brief.Subject + " do you mean?"
		if strings.TrimSpace(brief.Subject) == "" || len(question) > maxBytes-100 {
			question = "Which item do you mean? Try its name, tag, or location."
		}
		rows = []string{question}
		for _, finding := range brief.Findings {
			choice := "Possible match: " + finding.Title
			path := finding.ContainmentPath
			if len(path) > 0 && path[len(path)-1] == finding.Title {
				path = path[:len(path)-1]
			}
			if len(path) > 0 {
				choice += " in " + strings.Join(path, " / ")
			}
			rows = append(rows, choice+".")
		}
	default:
		if brief.Confidence == domain.ResponseConfidencePlausible || brief.Confidence == domain.ResponseConfidenceAmbiguous {
			prefix = "Possible matches for " + brief.Subject + ": "
		}
		for _, finding := range brief.Findings {
			row := finding.Title
			switch brief.Mode {
			case domain.ResponseAnswerModeLocate:
				path := finding.ContainmentPath
				if len(path) > 0 && path[len(path)-1] == finding.Title {
					path = path[:len(path)-1]
				}
				if len(path) == 0 {
					row += " has no recorded location."
				} else {
					row += " is in " + strings.Join(path, " / ") + "."
				}
			case domain.ResponseAnswerModeDetail, domain.ResponseAnswerModeHistory, domain.ResponseAnswerModeCheckout:
				if len(finding.Facts) == 0 {
					row += ": no further details recorded."
				} else {
					row += ": " + strings.Join(finding.Facts, "; ") + "."
				}
			default:
				row = "Found " + row + "."
			}
			rows = append(rows, row)
		}
	}
	const partial = " Showing only part of the results."
	text := prefix
	truncated := brief.Truncated
	for _, finding := range brief.Findings {
		truncated = truncated || finding.FactsTruncated
	}
	for _, row := range rows {
		next := strings.TrimSpace(text + " " + row)
		if len(next)+len(partial) > maxBytes {
			truncated = true
			break
		}
		text = next
	}
	if text == prefix && brief.Mode == domain.ResponseAnswerModeNotFound {
		text = "I couldn't find an item matching that description in the searched inventory results."
		truncated = brief.Truncated
	}
	if text == prefix {
		text = "I found results, but their details are too long to read here."
		if brief.Confidence == domain.ResponseConfidencePlausible || brief.Confidence == domain.ResponseConfidenceAmbiguous {
			text = "There are possible matches, but their details are too long to read here."
		}
		truncated = true
	}
	if truncated {
		text += partial
	}
	return strings.TrimSpace(text), nil
}
