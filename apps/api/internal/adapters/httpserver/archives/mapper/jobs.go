package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/archives/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func Job(j archivejob.Record) dto.ArchiveJob {
	return dto.ArchiveJob{ID: j.ID, Kind: string(j.Kind), State: string(j.State), Phase: string(j.Phase), InventoryID: j.SourceInventoryID, DestinationInventoryID: j.DestinationInventoryID, Photos: j.Photos, OtherFiles: j.OtherFiles, CreatedAt: j.CreatedAt, ExpiresAt: j.ExpiresAt, Failure: string(j.Failure)}
}
func Preview(p ports.ArchivePreview) dto.ArchivePreview {
	result := dto.ArchivePreview{InventoryName: p.InventoryName, Assets: p.Assets, Tags: p.Tags, CustomAssetTypes: p.CustomAssetTypes, CustomFields: p.CustomFields, Photos: p.Photos, OtherFiles: p.OtherFiles, OmittedAttachments: p.OmittedAttachments, KeyRemappings: []dto.ArchiveKeyRemapping{}}
	for _, r := range p.KeyRemappings {
		result.KeyRemappings = append(result.KeyRemappings, dto.ArchiveKeyRemapping{Family: r.Family, SourceKey: r.SourceKey, DestinationKey: r.DestinationKey})
	}
	return result
}
