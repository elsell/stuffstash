package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/imports/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
)

type ApplyCounts struct {
	FieldsCreated      int
	FieldsExisting     int
	LocationsCreated   int
	AssetsCreated      int
	AssetsSkipped      int
	AttachmentsCreated int
	AttachmentsSkipped int
}

func PreviewToResponse(plan importplan.Plan) dto.ImportPreviewResponse {
	counts := plan.Counts()
	return dto.ImportPreviewResponse{
		Source: dto.ImportSourceResponse{
			Type:        string(plan.Source.Type),
			Name:        plan.Source.Name,
			BaseURL:     plan.Source.BaseURL,
			Version:     plan.Source.Version,
			ImageImport: plan.Source.ImageImport,
		},
		Counts: dto.ImportCountsResponse{
			Fields:      counts.Fields,
			Locations:   counts.Locations,
			Assets:      counts.Assets,
			Attachments: counts.Attachments,
			Warnings:    counts.Warnings,
			Errors:      counts.Errors,
		},
		Fields:       fieldsToResponse(plan.Fields),
		AssetSamples: assetSamplesToResponse(plan.Assets, 25),
		ImageSamples: imageSamplesToResponse(plan.Attachments, 12),
		Messages:     messagesToResponse(plan.Messages),
	}
}

func ApplyToResponse(counts ApplyCounts, messages []importplan.Message) dto.ImportApplyResponse {
	return dto.ImportApplyResponse{
		Counts: dto.ImportApplyCountsResponse{
			FieldsCreated:      counts.FieldsCreated,
			FieldsExisting:     counts.FieldsExisting,
			LocationsCreated:   counts.LocationsCreated,
			AssetsCreated:      counts.AssetsCreated,
			AssetsSkipped:      counts.AssetsSkipped,
			AttachmentsCreated: counts.AttachmentsCreated,
			AttachmentsSkipped: counts.AttachmentsSkipped,
		},
		Messages: messagesToResponse(messages),
	}
}

func fieldsToResponse(fields []importplan.FieldDefinition) []dto.ImportFieldResponse {
	out := make([]dto.ImportFieldResponse, 0, len(fields))
	for _, field := range fields {
		out = append(out, dto.ImportFieldResponse{Key: field.Key, DisplayName: field.DisplayName, Type: field.Type})
	}
	return out
}

func assetSamplesToResponse(assets []importplan.Asset, limit int) []dto.ImportAssetSample {
	if len(assets) < limit {
		limit = len(assets)
	}
	out := make([]dto.ImportAssetSample, 0, limit)
	for _, asset := range assets[:limit] {
		out = append(out, dto.ImportAssetSample{
			SourceID:       asset.SourceID,
			Kind:           asset.Kind,
			Title:          asset.Title,
			Description:    asset.Description,
			ParentSourceID: asset.ParentSourceID,
			CustomFields:   asset.CustomFields,
		})
	}
	return out
}

func imageSamplesToResponse(attachments []importplan.Attachment, limit int) []dto.ImportImageSample {
	if len(attachments) < limit {
		limit = len(attachments)
	}
	out := make([]dto.ImportImageSample, 0, limit)
	for _, attachment := range attachments[:limit] {
		out = append(out, dto.ImportImageSample{
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

func messagesToResponse(messages []importplan.Message) []dto.ImportMessageResponse {
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
