package mapper

import (
	"slices"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/evaluationcases/dto"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func DefinitionToDomain(value dto.EvaluationCaseDefinition) domain.EvaluationCaseDefinitionInput {
	result := domain.EvaluationCaseDefinitionInput{Title: value.Title, Utterance: value.Utterance, Expectations: domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeKind(value.Expectations.Kind), ReferencedAssets: slices.Clone(value.Expectations.ReferencedAssets)}}
	for _, asset := range value.Assets {
		result.Assets = append(result.Assets, domain.EvaluationFixtureAsset{ID: asset.ID, Title: asset.Title, Kind: domain.EvaluationFixtureKind(asset.Kind), Description: asset.Description, ParentID: asset.ParentID, TagNames: slices.Clone(asset.TagNames)})
	}
	for _, location := range value.Expectations.Locations {
		result.Expectations.Locations = append(result.Expectations.Locations, domain.EvaluationLocationExpectation{AssetID: location.AssetID, AncestorID: location.AncestorID})
	}
	for _, proposal := range value.Expectations.Proposals {
		result.Expectations.Proposals = append(result.Expectations.Proposals, domain.EvaluationProposal{Operation: domain.Operation(proposal.Operation), TargetID: proposal.TargetID, DestinationID: proposal.DestinationID, NewTitle: proposal.NewTitle, NewKind: domain.EvaluationFixtureKind(proposal.NewKind), Details: proposal.Details})
	}
	for _, operation := range value.Expectations.ForbiddenOperations {
		result.Expectations.ForbiddenOperations = append(result.Expectations.ForbiddenOperations, domain.Operation(operation))
	}
	return result
}
func definitionToResponse(value domain.EvaluationCaseDefinitionInput) dto.EvaluationCaseDefinition {
	result := dto.EvaluationCaseDefinition{Title: value.Title, Utterance: value.Utterance, Expectations: dto.EvaluationCaseExpectations{Kind: string(value.Expectations.Kind), ReferencedAssets: slices.Clone(value.Expectations.ReferencedAssets)}}
	for _, asset := range value.Assets {
		result.Assets = append(result.Assets, dto.EvaluationCaseFixtureAsset{ID: asset.ID, Title: asset.Title, Kind: string(asset.Kind), Description: asset.Description, ParentID: asset.ParentID, TagNames: slices.Clone(asset.TagNames)})
	}
	for _, location := range value.Expectations.Locations {
		result.Expectations.Locations = append(result.Expectations.Locations, dto.EvaluationCaseLocation{AssetID: location.AssetID, AncestorID: location.AncestorID})
	}
	for _, proposal := range value.Expectations.Proposals {
		result.Expectations.Proposals = append(result.Expectations.Proposals, dto.EvaluationCaseProposal{Operation: string(proposal.Operation), TargetID: proposal.TargetID, DestinationID: proposal.DestinationID, NewTitle: proposal.NewTitle, NewKind: string(proposal.NewKind), Details: proposal.Details})
	}
	for _, operation := range value.Expectations.ForbiddenOperations {
		result.Expectations.ForbiddenOperations = append(result.Expectations.ForbiddenOperations, string(operation))
	}
	return result
}
func RevisionToResponse(revision domain.EvaluationCaseRevision) dto.EvaluationCaseRevision {
	value := revision.Snapshot()
	return dto.EvaluationCaseRevision{ID: string(value.ID), CaseID: string(value.CaseID), Number: value.Number, AuthorID: string(value.AuthorID), CreatedAt: value.CreatedAt, Definition: definitionToResponse(value.Definition.Settings())}
}
func HeadToResponse(value ports.EvaluationCaseHeadRecord) dto.EvaluationCaseHead {
	return dto.EvaluationCaseHead{ID: string(value.ID), Title: value.Title, LatestRevision: value.LatestRevision, LatestRevisionID: string(value.LatestRevisionID), CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
