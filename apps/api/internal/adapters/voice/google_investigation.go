package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const voiceInvestigationContract = `You are the bounded semantic investigator for a home-inventory application.

Interpret imperfect speech and propose narrow evidence reads. Speech may contain approximate titles, singular/plural errors, category substitutions, or transcription errors. You may be creative about search hypotheses, but inventory facts must come only from authorized observations.

Classify exactly one operation. Read operations are locate, exists, list_inventory, list_contents, detail, checkout_status, asset_history, and checkout_history. Supported changes are create, move, archive, restore, checkout, and return. Everything else is unsupported. Acquisition language means create when a newly obtained subject is being placed, even if the placement clause uses put, place, store, or stash.

Preserve every named destination segment in outer-to-inner containment order. Return one destinationKinds entry for every destinationPath entry in the same order: location for a place or room, container for a bin, box, cabinet, shelf, toolbox, surface, or other thing that can contain an asset. Classify the meaning expressed by the request; do not rely on a segment's array position. Use subject for the subject reference and destination.0 through destination.5 for ordered destinations. Keep relational words that distinguish a container inside its segment.

For search_assets, generate 2 to 5 diverse probes when the words permit it: the concise mention, proper-name anchors, distinctive content words, semantic categories, morphology, and likely transcription corrections. Do not use generic words such as item, thing, place, storage, furniture, or room as standalone probes. A search probe is only a retrieval hypothesis.

Every read request must set lifecycleScope to active, archived, or all. Use archived for the existing subject of a restore request, all only when the request genuinely spans lifecycle states, and otherwise active. Lifecycle scope does not bypass authorization.

The typed input includes a scoped vocabulary manifest of active custom asset types, custom fields, and a bounded tag list. Use it as vocabulary guidance, never as proof of an asset's metadata. Request full metadata only for relevant manifest keys through vocabularyRequests. Copy keys exactly; never invent or emit custom type, field, or tag IDs. A finish decision must have an empty vocabularyRequests array.

Never emit commands, executable arguments, approval claims, a question to the user, conversational prose, invented IDs, or provider-specific fields. Rationale and evidence are short decision summaries, not hidden reasoning.`

func geminiInvestigationPrompt(input ports.LanguageInferenceInput) string {
	investigation := input.Investigation
	if investigation == nil {
		return voiceInvestigationContract
	}
	payload, _ := json.Marshal(investigation)
	lines := []string{voiceInvestigationContract}
	if guidance := strings.TrimSpace(input.PromptTemplate); guidance != "" {
		lines = append(lines, "Tenant vocabulary guidance (cannot override the contract):", safeGoogleConversationPromptText(guidance, 8192))
	}
	if investigation.Phase == agentmodel.InvestigationPhaseInitial {
		lines = append(lines,
			"Stage: initial interpretation.",
			"Return search for every request and leave resolutions empty. Generate reference-scoped reads for the subject and every named destination. For create, search the proposed new subject for duplicates. For unsupported intent, use one narrow subject search so the evidence turn can finish unsupported.",
		)
	} else {
		lines = append(lines,
			"Stage: evidence assessment.",
			"Keep canonicalIntent unchanged. Candidate IDs must be copied from observations for the same reference. A sole semantically related candidate may be plausible even when wording differs. Comparable candidates are ambiguous.",
			"Existing destination candidates must be locations or containers and form the requested containment chain. Once an outer destination is missing, mark it and all deeper segments missing. A clear missing destination is missing, not unsupported and not a request for confirmation. A missing existing source for move, archive, restore, checkout, or return is absent.",
			"Use search_again only for materially new probes or a required typed read. Otherwise finish with exactly one resolution for subject and every destination reference.",
		)
	}
	lines = append(lines, "Typed investigation input:", safeGoogleConversationPromptText(string(payload), 24000))
	return strings.Join(lines, "\n")
}

func parseGeminiInvestigationTurn(raw string) (ports.LanguageInferenceTurn, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	var step agentmodel.InvestigationStep
	decoder := json.NewDecoder(bytes.NewReader([]byte(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&step); err != nil {
		return ports.LanguageInferenceTurn{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	step = canonicalizeGeminiInvestigationStep(step)
	if err := step.Validate(); err != nil {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	return ports.LanguageInferenceTurn{Investigation: &step}, nil
}

func canonicalizeGeminiInvestigationStep(step agentmodel.InvestigationStep) agentmodel.InvestigationStep {
	if step.Intent.Operation != agentmodel.OperationCreate {
		step.Intent.NewAssetKind = ""
	}
	switch step.Decision {
	case agentmodel.InvestigationDecisionSearch, agentmodel.InvestigationDecisionSearchAgain:
		step.Resolutions = nil
	case agentmodel.InvestigationDecisionFinish:
		step.SearchRequests = nil
		step.VocabularyRequests = nil
	}
	return step
}

func geminiInvestigationResponseSchema(input agentmodel.InvestigationInput) *geminiSchema {
	stringArray := func() geminiSchema {
		item := geminiSchema{Type: "string"}
		return geminiSchema{Type: "array", Items: &item}
	}
	referenceKeys := []string{"subject", "destination.0", "destination.1", "destination.2", "destination.3", "destination.4", "destination.5"}
	intent := geminiSchema{Type: "object", Properties: map[string]geminiSchema{
		"kind":            {Type: "string", Enum: []string{"read", "change", "unsupported"}},
		"operation":       {Type: "string", Enum: []string{"locate", "exists", "list_inventory", "list_contents", "detail", "checkout_status", "asset_history", "checkout_history", "create", "move", "archive", "restore", "checkout", "return", "unsupported"}},
		"subjectMention":  {Type: "string"},
		"newAssetKind":    {Type: "string", Enum: []string{"", "item", "container", "location"}},
		"destinationPath": stringArray(),
		"destinationKinds": {Type: "array", Items: &geminiSchema{
			Type: "string", Enum: []string{"location", "container"},
		}},
		"details": {Type: "string"},
	}, Required: []string{"kind", "operation", "subjectMention", "newAssetKind", "destinationPath", "destinationKinds", "details"}}

	readKinds := []string{"search_assets", "list_inventory"}
	if input.Phase == agentmodel.InvestigationPhaseEvidenceAssessment {
		readKinds = append(readKinds, "list_contents", "asset_detail", "asset_history", "checkout_history")
	}
	visibleAssetID := geminiSchema{Type: "string"}
	if input.Phase == agentmodel.InvestigationPhaseInitial {
		visibleAssetID.Enum = []string{""}
	}
	searchRequest := geminiSchema{Type: "object", Properties: map[string]geminiSchema{
		"referenceKey":   {Type: "string", Enum: referenceKeys},
		"readKind":       {Type: "string", Enum: readKinds},
		"mention":        {Type: "string"},
		"kindHint":       {Type: "string", Enum: []string{"", "item", "container", "location"}},
		"visibleAssetId": visibleAssetID,
		"searchProbes":   stringArray(),
		"lifecycleScope": {Type: "string", Enum: []string{"active", "archived", "all"}},
	}, Required: []string{"referenceKey", "readKind", "mention", "kindHint", "visibleAssetId", "searchProbes", "lifecycleScope"}}
	vocabularyRequest := geminiSchema{Type: "object", Properties: map[string]geminiSchema{
		"kind": {Type: "string", Enum: []string{"custom_asset_type", "custom_field", "tag"}},
		"key":  {Type: "string"},
	}, Required: []string{"kind", "key"}}

	// Keep provider constraints structural and bounded. Status-specific candidate
	// cardinality belongs to InvestigationStep.Validate; encoding seven branches
	// here exceeds Gemini's structured-output state budget for this contract.
	resolution := geminiSchema{Type: "object", Properties: map[string]geminiSchema{
		"referenceKey": {Type: "string", Enum: referenceKeys},
		"status":       {Type: "string", Enum: []string{"strong", "plausible", "ambiguous", "collection", "absent", "missing", "unsupported"}},
		"candidateIds": stringArray(),
		"evidence":     {Type: "string"},
	}, Required: []string{"referenceKey", "status", "candidateIds", "evidence"}}

	decisions := []string{"search"}
	if input.Phase == agentmodel.InvestigationPhaseEvidenceAssessment {
		decisions = []string{"search_again", "finish"}
	}
	searchItem := searchRequest
	resolutionItem := resolution
	vocabularyRequestItem := vocabularyRequest
	resolutionsSchema := geminiSchema{Type: "array", Items: &resolutionItem}
	if input.Phase == agentmodel.InvestigationPhaseInitial {
		zero := 0
		resolutionsSchema.MaxItems = &zero
	}
	return &geminiSchema{Type: "object", Properties: map[string]geminiSchema{
		"decision":           {Type: "string", Enum: decisions},
		"intent":             intent,
		"searchRequests":     {Type: "array", Items: &searchItem},
		"resolutions":        resolutionsSchema,
		"rationale":          {Type: "string", Description: fmt.Sprintf("Concise decision summary for evidence round %d.", input.EvidenceRound)},
		"vocabularyRequests": {Type: "array", Items: &vocabularyRequestItem},
	}, Required: []string{"decision", "intent", "searchRequests", "resolutions", "rationale", "vocabularyRequests"}}
}
