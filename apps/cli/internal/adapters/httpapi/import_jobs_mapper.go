package httpapi

import (
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func mapImportJob(v generated.ImportJobResponse) ports.ImportJob {
	var result ports.ImportJob
	if v.Actor != nil {
		item := mapImportJobActor(*v.Actor)
		result.Actor = &item
	}
	result.ActorID = v.ActorId
	result.CancellationMode = v.CancellationMode
	result.CompletedAt = v.CompletedAt
	result.Counts = mapImportJobCounts(v.Counts)
	result.CreatedAt = v.CreatedAt
	result.ID = v.Id
	if v.Messages.GetOrEmpty() != nil {
		result.Messages = make([]ports.ImportMessage, 0, len(v.Messages.GetOrEmpty()))
	}
	for _, item := range v.Messages.GetOrEmpty() {
		result.Messages = append(result.Messages, mapImportMessage(item))
	}
	result.Preview = mapImportJobPreview(v.Preview)
	result.Progress = mapImportJobProgress(v.Progress)
	if v.ProgressHistory.GetOrEmpty() != nil {
		result.ProgressHistory = make([]ports.ImportJobProgress, 0, len(v.ProgressHistory.GetOrEmpty()))
	}
	for _, item := range v.ProgressHistory.GetOrEmpty() {
		result.ProgressHistory = append(result.ProgressHistory, mapImportJobProgress(item))
	}
	if v.Resources.GetOrEmpty() != nil {
		result.Resources = make([]ports.ImportJobResource, 0, len(v.Resources.GetOrEmpty()))
	}
	for _, item := range v.Resources.GetOrEmpty() {
		result.Resources = append(result.Resources, mapImportJobResource(item))
	}
	result.Source = mapImportJobSource(v.Source)
	result.StartedAt = v.StartedAt
	result.Status = v.Status
	result.UpdatedAt = v.UpdatedAt
	return result
}

func mapImportJobActor(v generated.ImportJobActorResponse) ports.ImportJobActor {
	var result ports.ImportJobActor
	result.Email = v.Email
	result.ID = v.Id
	return result
}

func mapImportJobCounts(v generated.ImportJobCountsResponse) ports.ImportJobCounts {
	var result ports.ImportJobCounts
	result.Assets = v.Assets
	result.AssetsCreated = v.AssetsCreated
	result.AssetsSkipped = v.AssetsSkipped
	result.Attachments = v.Attachments
	result.AttachmentsCreated = v.AttachmentsCreated
	result.AttachmentsSkipped = v.AttachmentsSkipped
	result.Errors = v.Errors
	result.Fields = v.Fields
	result.FieldsCreated = v.FieldsCreated
	result.FieldsExisting = v.FieldsExisting
	result.Locations = v.Locations
	result.LocationsCreated = v.LocationsCreated
	result.RecordsDiscarded = v.RecordsDiscarded
	result.SourceLinksDiscarded = v.SourceLinksDiscarded
	result.Tags = v.Tags
	result.TagsCreated = v.TagsCreated
	result.TagsExisting = v.TagsExisting
	result.Warnings = v.Warnings
	return result
}

func mapImportMessage(v generated.ImportMessageResponse) ports.ImportMessage {
	var result ports.ImportMessage
	result.Code = v.Code
	result.Detail = v.Detail
	result.Severity = v.Severity
	result.SourceID = v.SourceId
	result.SourceName = v.SourceName
	result.Summary = v.Summary
	return result
}

func mapImportJobPreview(v generated.ImportJobPreview) ports.ImportJobPreview {
	var result ports.ImportJobPreview
	if v.Assets.GetOrEmpty() != nil {
		result.Assets = make([]ports.ImportJobPreviewAsset, 0, len(v.Assets.GetOrEmpty()))
	}
	for _, item := range v.Assets.GetOrEmpty() {
		result.Assets = append(result.Assets, mapImportJobPreviewAsset(item))
	}
	result.AssetsTruncated = v.AssetsTruncated
	if v.Attachments.GetOrEmpty() != nil {
		result.Attachments = make([]ports.ImportJobPreviewAttachment, 0, len(v.Attachments.GetOrEmpty()))
	}
	for _, item := range v.Attachments.GetOrEmpty() {
		result.Attachments = append(result.Attachments, mapImportJobPreviewAttachment(item))
	}
	result.AttachmentsTruncated = v.AttachmentsTruncated
	if v.Fields.GetOrEmpty() != nil {
		result.Fields = make([]ports.ImportJobPreviewField, 0, len(v.Fields.GetOrEmpty()))
	}
	for _, item := range v.Fields.GetOrEmpty() {
		result.Fields = append(result.Fields, mapImportJobPreviewField(item))
	}
	result.FieldsTruncated = v.FieldsTruncated
	if v.Locations.GetOrEmpty() != nil {
		result.Locations = make([]ports.ImportJobPreviewAsset, 0, len(v.Locations.GetOrEmpty()))
	}
	for _, item := range v.Locations.GetOrEmpty() {
		result.Locations = append(result.Locations, mapImportJobPreviewAsset(item))
	}
	result.LocationsTruncated = v.LocationsTruncated
	if v.Messages.GetOrEmpty() != nil {
		result.Messages = make([]ports.ImportMessage, 0, len(v.Messages.GetOrEmpty()))
	}
	for _, item := range v.Messages.GetOrEmpty() {
		result.Messages = append(result.Messages, mapImportMessage(item))
	}
	result.MessagesTruncated = v.MessagesTruncated
	if v.Tags.GetOrEmpty() != nil {
		result.Tags = make([]ports.ImportJobPreviewTag, 0, len(v.Tags.GetOrEmpty()))
	}
	for _, item := range v.Tags.GetOrEmpty() {
		result.Tags = append(result.Tags, mapImportJobPreviewTag(item))
	}
	result.TagsTruncated = v.TagsTruncated
	return result
}

func mapImportJobPreviewAsset(v generated.ImportJobPreviewAsset) ports.ImportJobPreviewAsset {
	var result ports.ImportJobPreviewAsset
	result.Archived = v.Archived
	result.Kind = v.Kind
	result.ParentSourceID = v.ParentSourceId
	result.SourceID = v.SourceId
	result.Title = v.Title
	return result
}

func mapImportJobPreviewAttachment(v generated.ImportJobPreviewAttachment) ports.ImportJobPreviewAttachment {
	var result ports.ImportJobPreviewAttachment
	result.AssetSourceID = v.AssetSourceId
	result.ContentType = v.ContentType
	result.FileName = v.FileName
	result.Primary = v.Primary
	result.SizeBytes = v.SizeBytes
	result.SourceID = v.SourceId
	return result
}

func mapImportJobPreviewField(v generated.ImportJobPreviewField) ports.ImportJobPreviewField {
	var result ports.ImportJobPreviewField
	result.DisplayName = v.DisplayName
	result.Key = v.Key
	result.Type = v.Type
	return result
}

func mapImportJobPreviewTag(v generated.ImportJobPreviewTag) ports.ImportJobPreviewTag {
	var result ports.ImportJobPreviewTag
	result.Color = v.Color
	result.DisplayName = v.DisplayName
	result.Key = v.Key
	return result
}

func mapImportJobProgress(v generated.ImportJobProgress) ports.ImportJobProgress {
	var result ports.ImportJobProgress
	result.Done = v.Done
	result.Message = v.Message
	result.Phase = v.Phase
	result.Total = v.Total
	result.UpdatedAt = v.UpdatedAt
	return result
}

func mapImportJobResource(v generated.ImportJobResource) ports.ImportJobResource {
	var result ports.ImportJobResource
	result.CreatedAt = v.CreatedAt
	result.DisplayName = v.DisplayName
	result.ResourceID = v.ResourceId
	result.ResourceOwnerID = v.ResourceOwnerId
	result.ResourceType = v.ResourceType
	result.SourceEntityID = v.SourceEntityId
	result.SourceEntityType = v.SourceEntityType
	return result
}

func mapImportJobSource(v generated.ImportJobSourceResponse) ports.ImportJobSource {
	var result ports.ImportJobSource
	result.AllowInsecureTLS = v.AllowInsecureTLS
	result.AllowPrivateNetwork = v.AllowPrivateNetwork
	result.BaseUrl = v.BaseUrl
	result.Fingerprint = v.Fingerprint
	result.ImageImport = v.ImageImport
	result.Name = v.Name
	result.Type = v.Type
	result.Version = v.Version
	return result
}
