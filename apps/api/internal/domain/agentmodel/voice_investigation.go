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
	MaxReadEvidenceRecords        = MaxSearchRequestsPerStep * MaxSearchProbesPerRequest * MaxEvidenceRounds
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

type RequestShape string

const (
	RequestShapeSingleTarget     RequestShape = "single_target"
	RequestShapeCollectionTarget RequestShape = "collection_target"
	RequestShapeCompound         RequestShape = "compound"
)

func (shape RequestShape) Valid() bool {
	switch shape {
	case RequestShapeSingleTarget, RequestShapeCollectionTarget, RequestShapeCompound:
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

func CanonicalizeIntent(intent Intent) Intent {
	shape := intent.RequestShape
	if shape == RequestShapeCompound || (shape == RequestShapeCollectionTarget && intent.Operation.changesInventory()) {
		intent.Operation = OperationUnsupported
	}
	switch {
	case intent.Operation.readOnly():
		intent.Kind = IntentKindRead
	case intent.Operation.changesInventory():
		intent.Kind = IntentKindChange
	case intent.Operation == OperationUnsupported:
		intent.Kind = IntentKindUnsupported
	default:
		intent.Kind = ""
	}
	if intent.Operation != OperationCreate {
		intent.NewAssetKind = ""
	}
	if intent.Operation != OperationCreate && intent.Operation != OperationMove {
		intent.DestinationPath = nil
		intent.DestinationKinds = nil
	}
	return intent
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

type LifecycleScope string

const (
	LifecycleScopeActive   LifecycleScope = "active"
	LifecycleScopeArchived LifecycleScope = "archived"
	LifecycleScopeAll      LifecycleScope = "all"
)

func (scope LifecycleScope) Valid() bool {
	return scope == "" || scope == LifecycleScopeActive || scope == LifecycleScopeArchived || scope == LifecycleScopeAll
}

func (scope LifecycleScope) Effective() LifecycleScope {
	if scope == "" {
		return LifecycleScopeActive
	}
	return scope
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

// DestinationKind is the user-facing containment role of one destination path
// segment. It is intentionally narrower than the asset kind enumeration: an
// item cannot contain another asset.
type DestinationKind string

const (
	DestinationKindLocation  DestinationKind = "location"
	DestinationKindContainer DestinationKind = "container"
)

func (kind DestinationKind) Valid() bool {
	return kind == DestinationKindLocation || kind == DestinationKindContainer
}

type Intent struct {
	RequestShape     RequestShape      `json:"requestShape"`
	Kind             IntentKind        `json:"kind"`
	Operation        Operation         `json:"operation"`
	SubjectMention   string            `json:"subjectMention"`
	NewAssetKind     string            `json:"newAssetKind"`
	DestinationPath  []string          `json:"destinationPath"`
	DestinationKinds []DestinationKind `json:"destinationKinds"`
	Details          string            `json:"details"`
}

func (intent Intent) Validate() error {
	if !intent.RequestShape.Valid() || !intent.Kind.Valid() || !intent.Operation.Valid() || !bounded(intent.SubjectMention, maxInvestigationTextRunes, true) ||
		!bounded(intent.Details, MaxInvestigationDetailRunes, true) || len(intent.DestinationPath) > MaxDestinationSegments ||
		len(intent.DestinationKinds) != len(intent.DestinationPath) {
		return ErrInvalidVoiceInvestigation
	}
	for index, segment := range intent.DestinationPath {
		if !bounded(segment, 200, false) || !intent.DestinationKinds[index].Valid() {
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
	shape := intent.RequestShape
	if (shape == RequestShapeCompound && intent.Operation != OperationUnsupported) ||
		(shape == RequestShapeCollectionTarget && intent.Operation.changesInventory()) {
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
	LifecycleScope LifecycleScope        `json:"lifecycleScope"`
}

func (request SearchRequest) Validate() error {
	if !request.ReferenceKey.Valid() || !request.ReadKind.Valid() || !bounded(request.Mention, 300, true) ||
		!bounded(request.VisibleAssetID, 200, true) || len(request.SearchProbes) > MaxSearchProbesPerRequest || !request.LifecycleScope.Valid() {
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
	ParentTitle     string               `json:"parentTitle,omitempty"`
	ParentKind      string               `json:"parentKind,omitempty"`
	LifecycleState  string               `json:"lifecycleState,omitempty"`
	CheckoutState   string               `json:"checkoutState,omitempty"`
	ContainmentPath []string             `json:"containmentPath,omitempty"`
	MatchedProbes   []string             `json:"matchedProbes,omitempty"`
	Facts           []string             `json:"facts,omitempty"`
}

func (observation CandidateObservation) Validate() error {
	if observation.EvidenceRound < 1 || observation.EvidenceRound > MaxEvidenceRounds || !observation.ReferenceKey.Valid() ||
		!bounded(observation.CandidateID, 200, false) || !bounded(observation.Title, 500, false) ||
		!bounded(observation.Description, maxInvestigationTextRunes, true) || !bounded(observation.ParentAssetID, 200, true) ||
		!bounded(observation.ParentTitle, 500, true) || !validCandidateObservationKind(observation.ParentKind, true) || len(observation.ContainmentPath) > 32 ||
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

func validCandidateObservationKind(value string, allowEmpty bool) bool {
	return (allowEmpty && value == "") || value == "item" || value == "container" || value == "location"
}

// ReadEvidence records one completed, authorization-scoped inventory read,
// including reads that returned no candidates.
type ReadEvidence struct {
	EvidenceRound  int                   `json:"evidenceRound"`
	ReferenceKey   SemanticReferenceKey  `json:"referenceKey"`
	ReadKind       InvestigationReadKind `json:"readKind"`
	Probe          string                `json:"probe,omitempty"`
	VisibleAssetID string                `json:"visibleAssetId,omitempty"`
	CandidateCount int                   `json:"candidateCount"`
	LifecycleScope LifecycleScope        `json:"lifecycleScope"`
}

func (evidence ReadEvidence) Validate() error {
	if evidence.EvidenceRound < 1 || evidence.EvidenceRound > MaxEvidenceRounds || !evidence.ReferenceKey.Valid() ||
		!evidence.ReadKind.Valid() || !bounded(evidence.Probe, 200, true) || !bounded(evidence.VisibleAssetID, 200, true) ||
		evidence.CandidateCount < 0 || evidence.CandidateCount > MaxCandidateObservations || !evidence.LifecycleScope.Valid() {
		return ErrInvalidVoiceInvestigation
	}
	switch evidence.ReadKind {
	case InvestigationReadSearchAssets:
		if strings.TrimSpace(evidence.Probe) == "" || strings.TrimSpace(evidence.VisibleAssetID) != "" {
			return ErrInvalidVoiceInvestigation
		}
	case InvestigationReadListInventory:
		if strings.TrimSpace(evidence.Probe) != "" || strings.TrimSpace(evidence.VisibleAssetID) != "" {
			return ErrInvalidVoiceInvestigation
		}
	default:
		if strings.TrimSpace(evidence.Probe) != "" || strings.TrimSpace(evidence.VisibleAssetID) == "" {
			return ErrInvalidVoiceInvestigation
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
	Phase                 InvestigationPhase          `json:"phase"`
	PromptVersion         string                      `json:"promptVersion"`
	SchemaVersion         string                      `json:"schemaVersion"`
	Transcript            string                      `json:"transcript"`
	EvidenceRound         int                         `json:"evidenceRound"`
	MaxEvidenceRounds     int                         `json:"maxEvidenceRounds"`
	CanonicalIntent       *Intent                     `json:"canonicalIntent,omitempty"`
	PreviousRequests      []SearchRequest             `json:"previousRequests"`
	Observations          []CandidateObservation      `json:"observations"`
	ReadEvidence          []ReadEvidence              `json:"readEvidence"`
	Vocabulary            VoiceVocabularyManifest     `json:"vocabulary"`
	VocabularyRequests    []VoiceVocabularyRequest    `json:"vocabularyRequests"`
	VocabularyDefinitions []VoiceVocabularyDefinition `json:"vocabularyDefinitions"`
}

func (input InvestigationInput) Validate() error {
	if !input.Phase.Valid() || !bounded(input.PromptVersion, 100, false) || !bounded(input.SchemaVersion, 100, false) ||
		!bounded(input.Transcript, maxInvestigationTextRunes, false) || input.MaxEvidenceRounds < 1 || input.MaxEvidenceRounds > MaxEvidenceRounds ||
		input.EvidenceRound < 0 || input.EvidenceRound > input.MaxEvidenceRounds || len(input.PreviousRequests) > MaxSearchRequestsPerStep*MaxEvidenceRounds ||
		len(input.Observations) > MaxCandidateObservations || len(input.ReadEvidence) > MaxReadEvidenceRecords || input.Vocabulary.Validate() != nil ||
		len(input.VocabularyRequests) > MaxVoiceVocabularyRequests || len(input.VocabularyDefinitions) > MaxVoiceVocabularyRequests {
		return ErrInvalidVoiceInvestigation
	}
	if input.Phase == InvestigationPhaseInitial {
		if input.EvidenceRound != 0 || input.CanonicalIntent != nil || len(input.PreviousRequests) != 0 || len(input.Observations) != 0 || len(input.ReadEvidence) != 0 || len(input.VocabularyRequests) != 0 || len(input.VocabularyDefinitions) != 0 {
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
	for _, evidence := range input.ReadEvidence {
		if evidence.Validate() != nil || evidence.EvidenceRound > input.EvidenceRound || !readEvidenceMatchesRequest(evidence, input.PreviousRequests) {
			return ErrInvalidVoiceInvestigation
		}
	}
	if !validVoiceVocabularyResolution(input.VocabularyRequests, input.VocabularyDefinitions) {
		return ErrInvalidVoiceInvestigation
	}
	for _, request := range input.PreviousRequests {
		if !requestCoveredByReadEvidence(request, input.ReadEvidence) {
			return ErrInvalidVoiceInvestigation
		}
	}
	return nil
}

func readEvidenceMatchesRequest(evidence ReadEvidence, requests []SearchRequest) bool {
	for _, request := range requests {
		if request.ReferenceKey != evidence.ReferenceKey || request.ReadKind != evidence.ReadKind || request.LifecycleScope.Effective() != evidence.LifecycleScope.Effective() {
			continue
		}
		if request.ReadKind == InvestigationReadSearchAssets {
			for _, probe := range request.SearchProbes {
				if strings.EqualFold(strings.TrimSpace(probe), strings.TrimSpace(evidence.Probe)) {
					return true
				}
			}
			continue
		}
		if request.ReadKind == InvestigationReadListInventory || strings.TrimSpace(request.VisibleAssetID) == strings.TrimSpace(evidence.VisibleAssetID) {
			return true
		}
	}
	return false
}

func requestCoveredByReadEvidence(request SearchRequest, records []ReadEvidence) bool {
	if request.ReadKind == InvestigationReadSearchAssets {
		for _, probe := range request.SearchProbes {
			covered := false
			for _, record := range records {
				if record.ReferenceKey == request.ReferenceKey && record.ReadKind == request.ReadKind && request.LifecycleScope.Effective() == record.LifecycleScope.Effective() && strings.EqualFold(strings.TrimSpace(record.Probe), strings.TrimSpace(probe)) {
					covered = true
					break
				}
			}
			if !covered {
				return false
			}
		}
		return true
	}
	for _, record := range records {
		if record.ReferenceKey == request.ReferenceKey && record.ReadKind == request.ReadKind && request.LifecycleScope.Effective() == record.LifecycleScope.Effective() &&
			(request.ReadKind == InvestigationReadListInventory || strings.TrimSpace(record.VisibleAssetID) == strings.TrimSpace(request.VisibleAssetID)) {
			return true
		}
	}
	return false
}

type InvestigationStep struct {
	Decision           InvestigationDecision    `json:"decision"`
	Intent             Intent                   `json:"intent"`
	SearchRequests     []SearchRequest          `json:"searchRequests"`
	Resolutions        []Resolution             `json:"resolutions"`
	Rationale          string                   `json:"rationale"`
	VocabularyRequests []VoiceVocabularyRequest `json:"vocabularyRequests"`
}

func (step InvestigationStep) Validate() error {
	if !step.Decision.Valid() || step.Intent.Validate() != nil || !bounded(step.Rationale, maxInvestigationEvidenceRunes, true) ||
		len(step.SearchRequests) > MaxSearchRequestsPerStep || len(step.Resolutions) > MaxDestinationSegments+1 || len(step.VocabularyRequests) > MaxVoiceVocabularyRequests {
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
	vocabularySeen := map[string]struct{}{}
	for _, request := range step.VocabularyRequests {
		if request.Validate() != nil {
			return ErrInvalidVoiceInvestigation
		}
		key := string(request.Kind) + "\x00" + request.Key
		if _, exists := vocabularySeen[key]; exists {
			return ErrInvalidVoiceInvestigation
		}
		vocabularySeen[key] = struct{}{}
	}
	switch step.Decision {
	case InvestigationDecisionSearch, InvestigationDecisionSearchAgain:
		if len(step.SearchRequests) == 0 || len(step.Resolutions) != 0 {
			return ErrInvalidVoiceInvestigation
		}
	case InvestigationDecisionFinish:
		if len(step.SearchRequests) != 0 || len(step.Resolutions) == 0 || len(step.VocabularyRequests) != 0 {
			return ErrInvalidVoiceInvestigation
		}
	}
	return nil
}

func validVoiceVocabularyResolution(requests []VoiceVocabularyRequest, definitions []VoiceVocabularyDefinition) bool {
	requested := map[string]struct{}{}
	for _, request := range requests {
		if request.Validate() != nil {
			return false
		}
		key := string(request.Kind) + "\x00" + request.Key
		if _, exists := requested[key]; exists {
			return false
		}
		requested[key] = struct{}{}
	}
	resolved := map[string]struct{}{}
	for _, definition := range definitions {
		if definition.Validate() != nil {
			return false
		}
		key := string(definition.Kind) + "\x00" + definition.Key
		if _, exists := requested[key]; !exists {
			return false
		}
		if _, exists := resolved[key]; exists {
			return false
		}
		resolved[key] = struct{}{}
	}
	return len(requested) == len(resolved)
}

func bounded(value string, limit int, optional bool) bool {
	trimmed := strings.TrimSpace(value)
	if !optional && trimmed == "" {
		return false
	}
	return utf8.RuneCountInString(trimmed) <= limit
}
