package mapper

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/imports/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
)

func JobListToResponse(jobs []importjob.Record) dto.ImportJobListResponse {
	out := make([]dto.ImportJobResponse, 0, len(jobs))
	for _, job := range jobs {
		out = append(out, JobToResponse(job))
	}
	return dto.ImportJobListResponse{Jobs: out}
}

func JobToResponse(job importjob.Record) dto.ImportJobResponse {
	return dto.ImportJobResponse{
		ID:      job.ID.String(),
		Status:  string(job.Status),
		ActorID: job.ActorID.String(),
		Source: dto.ImportJobSourceResponse{
			Type:                string(job.Source.Type),
			Name:                job.Source.Name,
			BaseURL:             job.Source.BaseURL,
			Version:             job.Source.Version,
			ImageImport:         job.Source.ImageImport,
			AllowPrivateNetwork: job.Source.AllowPrivateNetwork,
			AllowInsecureTLS:    job.Source.AllowInsecureTLS,
			Fingerprint:         job.Source.Fingerprint,
		},
		Counts: dto.ImportJobCountsResponse{
			Fields:               job.Counts.Fields,
			Locations:            job.Counts.Locations,
			Assets:               job.Counts.Assets,
			Attachments:          job.Counts.Attachments,
			Warnings:             job.Counts.Warnings,
			Errors:               job.Counts.Errors,
			FieldsCreated:        job.Counts.FieldsCreated,
			FieldsExisting:       job.Counts.FieldsExisting,
			LocationsCreated:     job.Counts.LocationsCreated,
			AssetsCreated:        job.Counts.AssetsCreated,
			AssetsSkipped:        job.Counts.AssetsSkipped,
			AttachmentsCreated:   job.Counts.AttachmentsCreated,
			AttachmentsSkipped:   job.Counts.AttachmentsSkipped,
			RecordsDiscarded:     job.Counts.RecordsDiscarded,
			SourceLinksDiscarded: job.Counts.SourceLinksDiscarded,
		},
		Preview: previewToResponse(job.Preview),
		Progress: dto.ImportJobProgress{
			Phase:     string(job.Progress.Phase),
			Done:      job.Progress.Done,
			Total:     job.Progress.Total,
			Message:   job.Progress.Message,
			UpdatedAt: timeString(job.Progress.UpdatedAt),
		},
		ProgressHistory:  progressHistoryToResponse(job.ProgressHistory),
		CancellationMode: string(job.CancellationMode),
		CreatedAt:        timeString(job.CreatedAt),
		StartedAt:        timeString(job.StartedAt),
		CompletedAt:      timeString(job.CompletedAt),
		UpdatedAt:        timeString(job.UpdatedAt),
		Resources:        resourcesToResponse(job.Resources),
		Messages:         messagesToResponse(job.Messages),
	}
}

func progressHistoryToResponse(history []importjob.Progress) []dto.ImportJobProgress {
	out := make([]dto.ImportJobProgress, 0, len(history))
	for _, progress := range history {
		out = append(out, dto.ImportJobProgress{
			Phase:     string(progress.Phase),
			Done:      progress.Done,
			Total:     progress.Total,
			Message:   progress.Message,
			UpdatedAt: timeString(progress.UpdatedAt),
		})
	}
	return out
}

func previewToResponse(preview importjob.PreviewSummary) dto.ImportJobPreview {
	return dto.ImportJobPreview{
		Fields:               previewFieldsToResponse(preview.Fields),
		Locations:            previewAssetsToResponse(preview.Locations),
		Assets:               previewAssetsToResponse(preview.Assets),
		Attachments:          previewAttachmentsToResponse(preview.Attachments),
		Messages:             messagesToResponse(preview.Messages),
		FieldsTruncated:      preview.FieldsTruncated,
		LocationsTruncated:   preview.LocationsTruncated,
		AssetsTruncated:      preview.AssetsTruncated,
		AttachmentsTruncated: preview.AttachmentsTruncated,
		MessagesTruncated:    preview.MessagesTruncated,
	}
}

func previewFieldsToResponse(fields []importjob.PreviewField) []dto.ImportJobPreviewField {
	out := make([]dto.ImportJobPreviewField, 0, len(fields))
	for _, field := range fields {
		out = append(out, dto.ImportJobPreviewField{
			Key:         field.Key,
			DisplayName: field.DisplayName,
			Type:        field.Type,
		})
	}
	return out
}

func previewAssetsToResponse(items []importjob.PreviewAsset) []dto.ImportJobPreviewAsset {
	out := make([]dto.ImportJobPreviewAsset, 0, len(items))
	for _, item := range items {
		out = append(out, dto.ImportJobPreviewAsset{
			SourceID:       item.SourceID,
			Kind:           item.Kind,
			Title:          item.Title,
			ParentSourceID: item.ParentSourceID,
			Archived:       item.Archived,
		})
	}
	return out
}

func previewAttachmentsToResponse(attachments []importjob.PreviewAttachment) []dto.ImportJobPreviewAttachment {
	out := make([]dto.ImportJobPreviewAttachment, 0, len(attachments))
	for _, attachment := range attachments {
		out = append(out, dto.ImportJobPreviewAttachment{
			SourceID:      attachment.SourceID,
			AssetSourceID: attachment.AssetSourceID,
			FileName:      attachment.FileName,
			ContentType:   attachment.ContentType,
			SizeBytes:     attachment.SizeBytes,
			Primary:       attachment.Primary,
		})
	}
	return out
}

func resourcesToResponse(resources []importjob.ResourceSummary) []dto.ImportJobResource {
	out := make([]dto.ImportJobResource, 0, len(resources))
	for _, resource := range resources {
		out = append(out, dto.ImportJobResource{
			ResourceType:     resource.ResourceType,
			ResourceID:       resource.ResourceID,
			DisplayName:      resource.DisplayName,
			ResourceOwnerID:  resource.ResourceOwnerID,
			SourceEntityType: resource.SourceEntityType,
			SourceEntityID:   resource.SourceEntityID,
			CreatedAt:        timeString(resource.CreatedAt),
		})
	}
	return out
}

func messagesToResponse(messages []importjob.Message) []dto.ImportMessageResponse {
	out := make([]dto.ImportMessageResponse, 0, len(messages))
	for _, message := range messages {
		out = append(out, dto.ImportMessageResponse{
			Code:       message.Code,
			Severity:   string(message.Severity),
			Summary:    message.Summary,
			Detail:     message.Detail,
			SourceID:   message.SourceID,
			SourceName: message.SourceName,
		})
	}
	return out
}

func timeString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
