package app

import (
	agentapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
)

type realtimeVoiceVocabularyCatalog = agentapp.VoiceVocabularyCatalog

func projectRealtimeVoiceVocabulary(assetTypes []customfield.AssetType, fields []customfield.Definition, tags []assettag.Tag) (agentmodel.VoiceVocabularyManifest, realtimeVoiceVocabularyCatalog, error) {
	return agentapp.ProjectVoiceVocabulary(assetTypes, fields, tags)
}
