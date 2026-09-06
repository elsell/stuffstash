package search

import (
	"strings"
)

type Mode string

const (
	ModeFuzzy Mode = "fuzzy"
	ModeExact Mode = "exact"
)

func NewMode(value string) (Mode, bool) {
	switch Mode(strings.TrimSpace(value)) {
	case "":
		return ModeFuzzy, true
	case ModeFuzzy:
		return ModeFuzzy, true
	case ModeExact:
		return ModeExact, true
	default:
		return "", false
	}
}

func (m Mode) String() string {
	return string(m)
}

type Query string

func NewQuery(value string) (Query, bool) {
	value = strings.TrimSpace(value)
	if len(value) > 120 {
		return "", false
	}
	return Query(value), true
}

func (q Query) String() string {
	return string(q)
}

// Terms preserves exact whole-value lookup and normalizes fuzzy conjunctions.
func (q Query) Terms(mode Mode) []string {
	value := strings.TrimSpace(strings.ToLower(q.String()))
	if value == "" {
		return nil
	}
	if mode == ModeExact {
		return []string{value}
	}
	seen := map[string]bool{}
	terms := []string{}
	for _, term := range strings.Fields(value) {
		if !seen[term] {
			seen[term] = true
			terms = append(terms, term)
		}
	}
	return terms
}

type ResultType string

const (
	ResultTypeAsset ResultType = "asset"
)

func (t ResultType) String() string {
	return string(t)
}

type MatchField string

const (
	MatchFieldTitle                 MatchField = "title"
	MatchFieldDescription           MatchField = "description"
	MatchFieldCustomField           MatchField = "custom_field"
	MatchFieldCustomAssetTypeKey    MatchField = "custom_asset_type_key"
	MatchFieldCustomAssetTypeName   MatchField = "custom_asset_type_name"
	MatchFieldCustomAssetTypeText   MatchField = "custom_asset_type_text"
	MatchFieldTagKey                MatchField = "tag_key"
	MatchFieldTagDisplayName        MatchField = "tag_display_name"
	MatchFieldAttachmentFileName    MatchField = "attachment_file_name"
	MatchFieldAttachmentContentType MatchField = "attachment_content_type"
)

func (f MatchField) String() string {
	return string(f)
}

type Match struct {
	Field MatchField
	Value string
}
