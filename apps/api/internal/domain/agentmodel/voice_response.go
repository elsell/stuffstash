package agentmodel

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var ErrInvalidGroundedVoiceResponseBrief = errors.New("invalid grounded voice response brief")

type ResponseBriefKind string

const (
	ResponseBriefKindAnswer        ResponseBriefKind = "answer"
	ResponseBriefKindClarification ResponseBriefKind = "clarification"
	ResponseBriefKindUnsupported   ResponseBriefKind = "unsupported_action"
)

func (kind ResponseBriefKind) Valid() bool {
	return kind == ResponseBriefKindAnswer || kind == ResponseBriefKindClarification || kind == ResponseBriefKindUnsupported
}

type ResponseAnswerMode string

const (
	ResponseAnswerModeLocate      ResponseAnswerMode = "locate"
	ResponseAnswerModeInventory   ResponseAnswerMode = "inventory"
	ResponseAnswerModeContents    ResponseAnswerMode = "contents"
	ResponseAnswerModeExists      ResponseAnswerMode = "exists"
	ResponseAnswerModeDetail      ResponseAnswerMode = "detail"
	ResponseAnswerModeHistory     ResponseAnswerMode = "history"
	ResponseAnswerModeCheckout    ResponseAnswerMode = "checkout"
	ResponseAnswerModeNotFound    ResponseAnswerMode = "not_found"
	ResponseAnswerModeClarify     ResponseAnswerMode = "clarify"
	ResponseAnswerModeUnsupported ResponseAnswerMode = "unsupported"
)

func (mode ResponseAnswerMode) Valid() bool {
	switch mode {
	case ResponseAnswerModeLocate, ResponseAnswerModeInventory, ResponseAnswerModeContents, ResponseAnswerModeExists,
		ResponseAnswerModeDetail, ResponseAnswerModeHistory, ResponseAnswerModeCheckout, ResponseAnswerModeNotFound,
		ResponseAnswerModeClarify, ResponseAnswerModeUnsupported:
		return true
	default:
		return false
	}
}

type ResponseConfidence string

const (
	ResponseConfidenceStrong    ResponseConfidence = "strong"
	ResponseConfidencePlausible ResponseConfidence = "plausible"
	ResponseConfidenceAmbiguous ResponseConfidence = "ambiguous"
	ResponseConfidenceAbsent    ResponseConfidence = "absent"
)

func (confidence ResponseConfidence) Valid() bool {
	return confidence == ResponseConfidenceStrong || confidence == ResponseConfidencePlausible || confidence == ResponseConfidenceAmbiguous || confidence == ResponseConfidenceAbsent
}

type ResponseFinding struct {
	FactKey         string   `json:"factKey"`
	Title           string   `json:"title"`
	Kind            string   `json:"kind"`
	LifecycleState  string   `json:"lifecycleState,omitempty"`
	CheckoutState   string   `json:"checkoutState,omitempty"`
	ContainmentPath []string `json:"containmentPath,omitempty"`
	Facts           []string `json:"facts,omitempty"`
	FactsTruncated  bool     `json:"factsTruncated,omitempty"`
}

type GroundedVoiceResponseBrief struct {
	Kind       ResponseBriefKind  `json:"kind"`
	Mode       ResponseAnswerMode `json:"mode"`
	Operation  Operation          `json:"operation"`
	Subject    string             `json:"subject"`
	Confidence ResponseConfidence `json:"confidence"`
	Findings   []ResponseFinding  `json:"findings,omitempty"`
	Truncated  bool               `json:"truncated,omitempty"`
}

func (brief GroundedVoiceResponseBrief) Validate() error {
	if !brief.Kind.Valid() || !brief.Mode.Valid() || !brief.Operation.Valid() || !brief.Confidence.Valid() || !boundedVoiceResponseText(brief.Subject, 160, true) || len(brief.Findings) > MaxCandidateObservations {
		return ErrInvalidGroundedVoiceResponseBrief
	}
	if brief.Kind == ResponseBriefKindClarification && brief.Mode != ResponseAnswerModeClarify {
		return ErrInvalidGroundedVoiceResponseBrief
	}
	if brief.Kind != ResponseBriefKindClarification && brief.Mode == ResponseAnswerModeClarify {
		return ErrInvalidGroundedVoiceResponseBrief
	}
	if (brief.Mode == ResponseAnswerModeLocate || brief.Mode == ResponseAnswerModeInventory || brief.Mode == ResponseAnswerModeContents || brief.Mode == ResponseAnswerModeExists || brief.Mode == ResponseAnswerModeDetail || brief.Mode == ResponseAnswerModeHistory || brief.Mode == ResponseAnswerModeCheckout || (brief.Mode == ResponseAnswerModeClarify && brief.Confidence != ResponseConfidenceAbsent)) && len(brief.Findings) == 0 {
		return ErrInvalidGroundedVoiceResponseBrief
	}
	seen := map[string]struct{}{}
	for _, finding := range brief.Findings {
		if !boundedVoiceResponseText(finding.FactKey, 40, false) || !boundedVoiceResponseText(finding.Title, 160, false) || !boundedVoiceResponseText(finding.Kind, 80, false) ||
			!validVoiceResponseLifecycleState(finding.LifecycleState) || !validVoiceResponseCheckoutState(finding.CheckoutState) || len(finding.ContainmentPath) > 3 || len(finding.Facts) > 3 {
			return ErrInvalidGroundedVoiceResponseBrief
		}
		if _, duplicate := seen[finding.FactKey]; duplicate {
			return ErrInvalidGroundedVoiceResponseBrief
		}
		seen[finding.FactKey] = struct{}{}
		for _, value := range append(append([]string{}, finding.ContainmentPath...), finding.Facts...) {
			if !boundedVoiceResponseText(value, 500, false) || internalVoiceResponseValue(value) {
				return ErrInvalidGroundedVoiceResponseBrief
			}
		}
	}
	return nil
}

func validVoiceResponseLifecycleState(value string) bool {
	return value == "" || value == "active" || value == "archived"
}

func validVoiceResponseCheckoutState(value string) bool {
	return value == "" || value == "available" || value == "checked_out"
}

func boundedVoiceResponseText(value string, limit int, allowEmpty bool) bool {
	value = strings.TrimSpace(value)
	return (allowEmpty || value != "") && utf8.ValidString(value) && utf8.RuneCountInString(value) <= limit && !internalVoiceResponseValue(value)
}

var internalVoiceResponsePattern = regexp.MustCompile(`(?i)\b(tenant|inventory|asset|parentasset|toolcall)[-_ ]?id\s*:`)

func internalVoiceResponseValue(value string) bool {
	return internalVoiceResponsePattern.MatchString(value)
}
