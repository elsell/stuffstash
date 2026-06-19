package search

import (
	"strings"
)

type AssetDocument struct {
	Title               string
	Description         string
	CustomFields        []string
	CustomAssetTypeKey  string
	CustomAssetTypeName string
	CustomAssetTypeText string
	Attachments         []AttachmentDocument
}

type AttachmentDocument struct {
	FileName    string
	ContentType string
}

func MatchAsset(document AssetDocument, query Query, mode Mode) []Match {
	matches := []Match{}
	if valueMatches(document.Title, query, mode) {
		matches = append(matches, Match{Field: MatchFieldTitle, Value: document.Title})
	}
	if valueMatches(document.Description, query, mode) {
		matches = append(matches, Match{Field: MatchFieldDescription, Value: document.Description})
	}
	for _, value := range document.CustomFields {
		if valueMatches(value, query, mode) {
			matches = append(matches, Match{Field: MatchFieldCustomField, Value: value})
		}
	}
	if document.CustomAssetTypeKey != "" || document.CustomAssetTypeName != "" || document.CustomAssetTypeText != "" {
		for _, candidate := range []struct {
			field MatchField
			value string
		}{
			{field: MatchFieldCustomAssetTypeKey, value: document.CustomAssetTypeKey},
			{field: MatchFieldCustomAssetTypeName, value: document.CustomAssetTypeName},
			{field: MatchFieldCustomAssetTypeText, value: document.CustomAssetTypeText},
		} {
			if valueMatches(candidate.value, query, mode) {
				matches = append(matches, Match{Field: candidate.field, Value: candidate.value})
			}
		}
	}
	for _, attachment := range document.Attachments {
		for _, candidate := range []struct {
			field MatchField
			value string
		}{
			{field: MatchFieldAttachmentFileName, value: attachment.FileName},
			{field: MatchFieldAttachmentContentType, value: attachment.ContentType},
		} {
			if valueMatches(candidate.value, query, mode) {
				matches = append(matches, Match{Field: candidate.field, Value: candidate.value})
			}
		}
	}
	return matches
}

func valueMatches(value string, query Query, mode Mode) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	queryValue := strings.TrimSpace(strings.ToLower(query.String()))
	if value == "" || queryValue == "" {
		return false
	}
	switch mode {
	case ModeExact:
		return value == queryValue
	default:
		return strings.Contains(value, queryValue)
	}
}
