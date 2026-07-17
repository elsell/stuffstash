package agentmodel

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	MaxEvidenceRounds             = 2
	MaxDestinationSegments        = 6
	MaxSearchProbesPerRequest     = 5
	MaxSearchRequestsPerStep      = 24
	MaxCandidateObservations      = 120
	MaxObservationFacts           = 12
	MaxInvestigationDetailRunes   = 500
	maxInvestigationTextRunes     = 1000
	maxInvestigationEvidenceRunes = 500
)

var ErrInvalidVoiceInvestigation = errors.New("invalid voice investigation")

type IntentKind string

const (
	IntentKindRead        IntentKind = "read"
	IntentKindChange      IntentKind = "change"
	IntentKindUnsupported IntentKind = "unsupported"
)

func (kind IntentKind) Valid() bool {
	switch kind {
	case IntentKindRead, IntentKindChange, IntentKindUnsupported:
		return true
	default:
		return false
	}
}

type Operation string

const (
	OperationLocate          Operation = "locate"
	OperationExists          Operation = "exists"
	OperationListInventory   Operation = "list_inventory"
	OperationListContents    Operation = "list_contents"
	OperationDetail          Operation = "detail"
	OperationCheckoutStatus  Operation = "checkout_status"
	OperationAssetHistory    Operation = "asset_history"
	OperationCheckoutHistory Operation = "checkout_history"
	OperationCreate          Operation = "create"
	OperationMove            Operation = "move"
	OperationArchive         Operation = "archive"
	OperationRestore         Operation = "restore"
	OperationCheckout        Operation = "checkout"
	OperationReturn          Operation = "return"
	OperationUnsupported     Operation = "unsupported"
)

func (operation Operation) Valid() bool {
	switch operation {
	case OperationLocate, OperationExists, OperationListInventory, OperationListContents,
		OperationDetail, OperationCheckoutStatus, OperationAssetHistory, OperationCheckoutHistory,
		OperationCreate, OperationMove, OperationArchive, OperationRestore, OperationCheckout,
		OperationReturn, OperationUnsupported:
		return true
	default:
		return false
	}
}

func (operation Operation) readOnly() bool {
	switch operation {
	case OperationLocate, OperationExists, OperationListInventory, OperationListContents,
		OperationDetail, OperationCheckoutStatus, OperationAssetHistory, OperationCheckoutHistory:
		return true
	default:
		return false
	}
}

func (operation Operation) changesInventory() bool {
	switch operation {
	case OperationCreate, OperationMove, OperationArchive, OperationRestore, OperationCheckout, OperationReturn:
		return true
	default:
		return false
	}
}

type InvestigationDecision string

const (
	InvestigationDecisionSearch      InvestigationDecision = "search"
	InvestigationDecisionSearchAgain InvestigationDecision = "search_again"
	InvestigationDecisionFinish      InvestigationDecision = "finish"
)

func (decision InvestigationDecision) Valid() bool {
	switch decision {
	case InvestigationDecisionSearch, InvestigationDecisionSearchAgain, InvestigationDecisionFinish:
		return true
	default:
		return false
	}
}

type InvestigationPhase string

const (
	InvestigationPhaseInitial            InvestigationPhase = "initial"
	InvestigationPhaseEvidenceAssessment InvestigationPhase = "evidence_assessment"
)

func (phase InvestigationPhase) Valid() bool {
	return phase == InvestigationPhaseInitial || phase == InvestigationPhaseEvidenceAssessment
}

type InvestigationReadKind string

const (
	InvestigationReadSearchAssets    InvestigationReadKind = "search_assets"
	InvestigationReadListInventory   InvestigationReadKind = "list_inventory"
	InvestigationReadListContents    InvestigationReadKind = "list_contents"
	InvestigationReadAssetDetail     InvestigationReadKind = "asset_detail"
	InvestigationReadAssetHistory    InvestigationReadKind = "asset_history"
	InvestigationReadCheckoutHistory InvestigationReadKind = "checkout_history"
)

func (kind InvestigationReadKind) Valid() bool {
	switch kind {
	case InvestigationReadSearchAssets, InvestigationReadListInventory, InvestigationReadListContents,
		InvestigationReadAssetDetail, InvestigationReadAssetHistory, InvestigationReadCheckoutHistory:
		return true
	default:
		return false
	}
}

type ResolutionStatus string

const (
	ResolutionStrong      ResolutionStatus = "strong"
	ResolutionPlausible   ResolutionStatus = "plausible"
	ResolutionAmbiguous   ResolutionStatus = "ambiguous"
	ResolutionCollection  ResolutionStatus = "collection"
	ResolutionAbsent      ResolutionStatus = "absent"
	ResolutionMissing     ResolutionStatus = "missing"
	ResolutionUnsupported ResolutionStatus = "unsupported"
)

func (status ResolutionStatus) Valid() bool {
	switch status {
	case ResolutionStrong, ResolutionPlausible, ResolutionAmbiguous, ResolutionCollection,
		ResolutionAbsent, ResolutionMissing, ResolutionUnsupported:
		return true
	default:
		return false
	}
}

type SemanticReferenceKey string

const SemanticReferenceSubject SemanticReferenceKey = "subject"

func NewSemanticReferenceKey(value string) (SemanticReferenceKey, bool) {
	value = strings.TrimSpace(value)
	if value == string(SemanticReferenceSubject) {
		return SemanticReferenceSubject, true
	}
	const prefix = "destination."
	if !strings.HasPrefix(value, prefix) {
		return "", false
	}
	rawIndex := strings.TrimPrefix(value, prefix)
	if rawIndex == "" || (len(rawIndex) > 1 && rawIndex[0] == '0') {
		return "", false
	}
	index, err := strconv.Atoi(rawIndex)
	if err != nil || index < 0 || index >= MaxDestinationSegments {
		return "", false
	}
	return SemanticReferenceKey(value), true
}

func (key SemanticReferenceKey) String() string { return string(key) }

func (key SemanticReferenceKey) Valid() bool {
	_, ok := NewSemanticReferenceKey(string(key))
	return ok
}

type Intent struct {
	Kind            IntentKind `json:"kind"`
	Operation       Operation  `json:"operation"`
	SubjectMention  string     `json:"subjectMention"`
	NewAssetKind    string     `json:"newAssetKind"`
	DestinationPath []string   `json:"destinationPath"`
	Details         string     `json:"details"`
}

func (intent Intent) Validate() error {
	if !intent.Kind.Valid() || !intent.Operation.Valid() || !bounded(intent.SubjectMention, maxInvestigationTextRunes, true) ||
		!bounded(intent.Details, MaxInvestigationDetailRunes, true) || len(intent.DestinationPath) > MaxDestinationSegments {
		return ErrInvalidVoiceInvestigation
	}
	for _, segment := range intent.DestinationPath {
		if !bounded(segment, 200, false) {
			return ErrInvalidVoiceInvestigation
		}
	}
	switch intent.Kind {
	case IntentKindRead:
		if !intent.Operation.readOnly() || (intent.Operation != OperationListInventory && strings.TrimSpace(intent.SubjectMention) == "") {
			return ErrInvalidVoiceInvestigation
		}
	case IntentKindChange:
		if !intent.Operation.changesInventory() || strings.TrimSpace(intent.SubjectMention) == "" {
			return ErrInvalidVoiceInvestigation
		}
	case IntentKindUnsupported:
		if intent.Operation != OperationUnsupported {
			return ErrInvalidVoiceInvestigation
		}
	}
	if intent.Operation == OperationMove && len(intent.DestinationPath) == 0 {
		return ErrInvalidVoiceInvestigation
	}
	if intent.Operation == OperationCreate {
		switch intent.NewAssetKind {
		case "item", "container", "location":
		default:
			return ErrInvalidVoiceInvestigation
		}
	} else if strings.TrimSpace(intent.NewAssetKind) != "" {
		return ErrInvalidVoiceInvestigation
	}
	return nil
}

type SearchRequest struct {
	ReferenceKey   SemanticReferenceKey  `json:"referenceKey"`
	ReadKind       InvestigationReadKind `json:"readKind"`
	Mention        string                `json:"mention"`
	KindHint       string                `json:"kindHint"`
	VisibleAssetID string                `json:"visibleAssetId"`
	SearchProbes   []string              `json:"searchProbes"`
}

func (request SearchRequest) Validate() error {
	if !request.ReferenceKey.Valid() || !request.ReadKind.Valid() || !bounded(request.Mention, 300, true) ||
		!bounded(request.VisibleAssetID, 200, true) || len(request.SearchProbes) > MaxSearchProbesPerRequest {
		return ErrInvalidVoiceInvestigation
	}
	if request.KindHint != "" && request.KindHint != "item" && request.KindHint != "container" && request.KindHint != "location" {
		return ErrInvalidVoiceInvestigation
	}
	seen := map[string]struct{}{}
	for _, probe := range request.SearchProbes {
		if !bounded(probe, 200, false) {
			return ErrInvalidVoiceInvestigation
		}
		key := strings.ToLower(strings.Join(strings.Fields(probe), " "))
		if _, exists := seen[key]; exists {
			return ErrInvalidVoiceInvestigation
		}
		seen[key] = struct{}{}
	}
	switch request.ReadKind {
	case InvestigationReadSearchAssets:
		if strings.TrimSpace(request.Mention) == "" || len(request.SearchProbes) == 0 || request.VisibleAssetID != "" {
			return ErrInvalidVoiceInvestigation
		}
	case InvestigationReadListInventory:
		if request.VisibleAssetID != "" || len(request.SearchProbes) != 0 {
			return ErrInvalidVoiceInvestigation
		}
	default:
		if strings.TrimSpace(request.VisibleAssetID) == "" || len(request.SearchProbes) != 0 {
			return ErrInvalidVoiceInvestigation
		}
	}
	return nil
}

type CandidateObservation struct {
	EvidenceRound   int                  `json:"evidenceRound"`
	ReferenceKey    SemanticReferenceKey `json:"referenceKey"`
	CandidateID     string               `json:"candidateId"`
	Title           string               `json:"title"`
	Kind            string               `json:"kind"`
	Description     string               `json:"description,omitempty"`
	ParentAssetID   string               `json:"parentAssetId,omitempty"`
	LifecycleState  string               `json:"lifecycleState,omitempty"`
	CheckoutState   string               `json:"checkoutState,omitempty"`
	ContainmentPath []string             `json:"containmentPath,omitempty"`
	MatchedProbes   []string             `json:"matchedProbes,omitempty"`
	Facts           []string             `json:"facts,omitempty"`
}

func (observation CandidateObservation) Validate() error {
	if observation.EvidenceRound < 1 || observation.EvidenceRound > MaxEvidenceRounds || !observation.ReferenceKey.Valid() ||
		!bounded(observation.CandidateID, 200, false) || !bounded(observation.Title, 500, false) ||
		!bounded(observation.Description, maxInvestigationTextRunes, true) || len(observation.ContainmentPath) > 32 ||
		len(observation.MatchedProbes) > MaxSearchProbesPerRequest || len(observation.Facts) > MaxObservationFacts {
		return ErrInvalidVoiceInvestigation
	}
	for _, values := range [][]string{observation.ContainmentPath, observation.MatchedProbes, observation.Facts} {
		for _, value := range values {
			if !bounded(value, 500, false) {
				return ErrInvalidVoiceInvestigation
			}
		}
	}
	return nil
}

type Resolution struct {
	ReferenceKey SemanticReferenceKey `json:"referenceKey"`
	Status       ResolutionStatus     `json:"status"`
	CandidateIDs []string             `json:"candidateIds"`
	Evidence     string               `json:"evidence"`
}

func (resolution Resolution) Validate() error {
	if !resolution.ReferenceKey.Valid() || !resolution.Status.Valid() || !bounded(resolution.Evidence, maxInvestigationEvidenceRunes, true) {
		return ErrInvalidVoiceInvestigation
	}
	seen := map[string]struct{}{}
	for _, id := range resolution.CandidateIDs {
		if !bounded(id, 200, false) {
			return ErrInvalidVoiceInvestigation
		}
		if _, exists := seen[id]; exists {
			return ErrInvalidVoiceInvestigation
		}
		seen[id] = struct{}{}
	}
	switch resolution.Status {
	case ResolutionStrong, ResolutionPlausible:
		if len(resolution.CandidateIDs) != 1 {
			return ErrInvalidVoiceInvestigation
		}
	case ResolutionAmbiguous:
		if len(resolution.CandidateIDs) < 2 {
			return ErrInvalidVoiceInvestigation
		}
	case ResolutionAbsent, ResolutionMissing, ResolutionUnsupported:
		if len(resolution.CandidateIDs) != 0 {
			return ErrInvalidVoiceInvestigation
		}
	}
	return nil
}

type InvestigationInput struct {
	Phase             InvestigationPhase     `json:"phase"`
	PromptVersion     string                 `json:"promptVersion"`
	SchemaVersion     string                 `json:"schemaVersion"`
	Transcript        string                 `json:"transcript"`
	EvidenceRound     int                    `json:"evidenceRound"`
	MaxEvidenceRounds int                    `json:"maxEvidenceRounds"`
	CanonicalIntent   *Intent                `json:"canonicalIntent,omitempty"`
	PreviousRequests  []SearchRequest        `json:"previousRequests"`
	Observations      []CandidateObservation `json:"observations"`
}

func (input InvestigationInput) Validate() error {
	if !input.Phase.Valid() || !bounded(input.PromptVersion, 100, false) || !bounded(input.SchemaVersion, 100, false) ||
		!bounded(input.Transcript, maxInvestigationTextRunes, false) || input.MaxEvidenceRounds < 1 || input.MaxEvidenceRounds > MaxEvidenceRounds ||
		input.EvidenceRound < 0 || input.EvidenceRound > input.MaxEvidenceRounds || len(input.PreviousRequests) > MaxSearchRequestsPerStep*MaxEvidenceRounds ||
		len(input.Observations) > MaxCandidateObservations {
		return ErrInvalidVoiceInvestigation
	}
	if input.Phase == InvestigationPhaseInitial {
		if input.EvidenceRound != 0 || input.CanonicalIntent != nil || len(input.PreviousRequests) != 0 || len(input.Observations) != 0 {
			return ErrInvalidVoiceInvestigation
		}
	} else {
		if input.EvidenceRound < 1 || input.CanonicalIntent == nil || input.CanonicalIntent.Validate() != nil || len(input.PreviousRequests) == 0 {
			return ErrInvalidVoiceInvestigation
		}
	}
	for _, request := range input.PreviousRequests {
		if request.Validate() != nil {
			return ErrInvalidVoiceInvestigation
		}
	}
	for _, observation := range input.Observations {
		if observation.Validate() != nil || observation.EvidenceRound > input.EvidenceRound {
			return ErrInvalidVoiceInvestigation
		}
	}
	return nil
}

type InvestigationStep struct {
	Decision       InvestigationDecision `json:"decision"`
	Intent         Intent                `json:"intent"`
	SearchRequests []SearchRequest       `json:"searchRequests"`
	Resolutions    []Resolution          `json:"resolutions"`
	Rationale      string                `json:"rationale"`
}

func (step InvestigationStep) Validate() error {
	if !step.Decision.Valid() || step.Intent.Validate() != nil || !bounded(step.Rationale, maxInvestigationEvidenceRunes, true) ||
		len(step.SearchRequests) > MaxSearchRequestsPerStep || len(step.Resolutions) > MaxDestinationSegments+1 {
		return ErrInvalidVoiceInvestigation
	}
	for _, request := range step.SearchRequests {
		if request.Validate() != nil {
			return ErrInvalidVoiceInvestigation
		}
	}
	seen := map[SemanticReferenceKey]struct{}{}
	for _, resolution := range step.Resolutions {
		if resolution.Validate() != nil {
			return ErrInvalidVoiceInvestigation
		}
		if _, exists := seen[resolution.ReferenceKey]; exists {
			return ErrInvalidVoiceInvestigation
		}
		seen[resolution.ReferenceKey] = struct{}{}
	}
	switch step.Decision {
	case InvestigationDecisionSearch, InvestigationDecisionSearchAgain:
		if len(step.SearchRequests) == 0 || len(step.Resolutions) != 0 {
			return ErrInvalidVoiceInvestigation
		}
	case InvestigationDecisionFinish:
		if len(step.SearchRequests) != 0 || len(step.Resolutions) == 0 {
			return ErrInvalidVoiceInvestigation
		}
	}
	return nil
}

func bounded(value string, limit int, optional bool) bool {
	trimmed := strings.TrimSpace(value)
	if !optional && trimmed == "" {
		return false
	}
	return utf8.RuneCountInString(trimmed) <= limit
}
