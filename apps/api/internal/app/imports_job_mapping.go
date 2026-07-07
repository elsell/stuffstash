package app

import (
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func importJobTenantID(id importjob.TenantID) tenant.ID {
	return tenant.ID(id.String())
}

func importJobInventoryID(id importjob.InventoryID) inventory.InventoryID {
	return inventory.InventoryID(id.String())
}

func importJobPrincipalID(id importjob.PrincipalID) identity.PrincipalID {
	return identity.PrincipalID(id.String())
}

func importJobCountsFromPlan(plan importplan.Plan) importjob.Counts {
	counts := plan.Counts()
	jobCounts := importjob.Counts{
		Fields:      counts.Fields,
		Locations:   counts.Locations,
		Assets:      counts.Assets,
		Attachments: counts.Attachments,
		Warnings:    counts.Warnings,
		Errors:      counts.Errors,
	}
	for _, message := range plan.Messages {
		switch message.Code {
		case "duplicate-source-asset", "archived-source-asset-skipped":
			jobCounts.AssetsSkipped++
		case "duplicate-source-attachment":
			jobCounts.AttachmentsSkipped++
		}
	}
	return jobCounts
}

func importJobPreviewSummaryFromPlan(plan importplan.Plan, limit int) importjob.PreviewSummary {
	if limit <= 0 {
		limit = 12
	}
	summary := importjob.PreviewSummary{
		FieldsTruncated:      len(plan.Fields) > limit,
		AttachmentsTruncated: len(plan.Attachments) > limit,
		MessagesTruncated:    len(plan.Messages) > limit,
	}
	for _, field := range plan.Fields {
		if len(summary.Fields) >= limit {
			break
		}
		summary.Fields = append(summary.Fields, importjob.PreviewField{
			Key:         field.Key,
			DisplayName: field.DisplayName,
			Type:        field.Type,
		})
	}
	for _, item := range plan.Assets {
		if item.Kind != "location" {
			continue
		}
		if len(summary.Locations) >= limit {
			summary.LocationsTruncated = true
			continue
		}
		summary.Locations = append(summary.Locations, importjob.PreviewAsset{
			SourceID:       item.SourceID,
			Kind:           item.Kind,
			Title:          item.Title,
			ParentSourceID: item.ParentSourceID,
			Archived:       item.Archived,
		})
	}
	for _, item := range plan.Assets {
		if item.Kind == "location" {
			continue
		}
		if len(summary.Assets) >= limit {
			summary.AssetsTruncated = true
			continue
		}
		summary.Assets = append(summary.Assets, importjob.PreviewAsset{
			SourceID:       item.SourceID,
			Kind:           item.Kind,
			Title:          item.Title,
			ParentSourceID: item.ParentSourceID,
			Archived:       item.Archived,
		})
	}
	for _, attachment := range plan.Attachments {
		if len(summary.Attachments) >= limit {
			break
		}
		summary.Attachments = append(summary.Attachments, importjob.PreviewAttachment{
			SourceID:      attachment.SourceID,
			AssetSourceID: attachment.AssetSourceID,
			FileName:      attachment.FileName,
			ContentType:   attachment.ContentType,
			SizeBytes:     attachment.SizeBytes,
			Primary:       attachment.Primary,
		})
	}
	for _, message := range importJobMessagesFromPlanMessages(plan.Messages) {
		if len(summary.Messages) >= limit {
			break
		}
		summary.Messages = append(summary.Messages, message)
	}
	return summary
}

func importJobSourceRefFromPlan(plan importplan.Plan, fingerprint string, requests ...ports.ImportSourceRequest) importjob.SourceRef {
	ref := importjob.SourceRef{
		Type:        importjob.SourceType(plan.Source.Type),
		Name:        plan.Source.Name,
		BaseURL:     plan.Source.BaseURL,
		Version:     plan.Source.Version,
		ImageImport: plan.Source.ImageImport,
		Fingerprint: fingerprint,
	}
	if len(requests) > 0 {
		ref.AllowPrivateNetwork = requests[0].AllowPrivateNetwork
		ref.AllowInsecureTLS = requests[0].AllowInsecureTLS
	}
	return ref
}

func importJobMessagesFromPlanMessages(messages []importplan.Message) []importjob.Message {
	out := make([]importjob.Message, 0, len(messages))
	for _, message := range messages {
		out = append(out, importJobMessageFromPlanMessage(message))
	}
	return out
}

func importJobMessageFromPlanMessage(message importplan.Message) importjob.Message {
	return importjob.Message{
		Code:       message.Code,
		Severity:   importjob.MessageSeverity(message.Severity),
		Summary:    message.Summary,
		Detail:     message.Detail,
		SourceID:   message.SourceID,
		SourceName: message.SourceName,
	}
}
