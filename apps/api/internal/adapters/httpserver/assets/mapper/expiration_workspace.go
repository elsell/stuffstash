package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/assets/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func ExpirationWorkspaceAssetsToResponse(records []ports.AssetSearchResult, photos map[ports.AttachmentAssetReference]media.Attachment) []dto.ExpirationWorkspaceAsset {
	items := []dto.ExpirationWorkspaceAsset{}
	for _, record := range records {
		checkouts := map[asset.ID]asset.Checkout{}
		if record.CurrentCheckout != nil {
			checkouts[record.Asset.ID] = *record.CurrentCheckout
		}
		mapped := AssetsToResponseWithTags([]asset.Asset{record.Asset}, map[asset.ID][]assettag.Tag{record.Asset.ID: record.AssignedTags}, photos, checkouts, nil)
		item := dto.ExpirationWorkspaceAsset{AssetResponse: mapped[0], AncestorPath: []dto.ExpirationWorkspaceAncestor{}}
		for _, parent := range record.AncestorPath {
			item.AncestorPath = append(item.AncestorPath, dto.ExpirationWorkspaceAncestor{ID: parent.ID.String(), Title: parent.Title.String()})
		}
		items = append(items, item)
	}
	return items
}

// ExpirationWorkspacePresentation keeps application values out of transport mapping.
type ExpirationWorkspacePresentation struct {
	Soon, Expired, All int
	Timezone           string
	Contexts           []ExpirationWorkspaceContext
}
type ExpirationWorkspaceContext struct {
	State           expirationdate.State
	TrackingEnabled bool
	AdvanceDays     int
	Timezone        string
}

func ExpirationWorkspaceToResponse(records []ports.AssetSearchResult, photos map[ports.AttachmentAssetReference]media.Attachment, presentation ExpirationWorkspacePresentation) dto.ExpirationWorkspaceData {
	items := ExpirationWorkspaceAssetsToResponse(records, photos)
	for index, value := range presentation.Contexts {
		items[index].ExpirationContext = ExpirationContextToResponse(value.State, value.TrackingEnabled, value.AdvanceDays, value.Timezone)
	}
	return dto.ExpirationWorkspaceData{Items: items, Counts: dto.ExpirationWorkspaceCounts{Soon: presentation.Soon, Expired: presentation.Expired, All: presentation.All}, Timezone: presentation.Timezone}
}
