package inventoryarchive

import (
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func projectmetadataTag(v assettag.Tag) metadataTag {
	return metadataTag{ID: v.ID, Key: v.Key, DisplayName: v.DisplayName, Color: v.Color, LifecycleState: v.LifecycleState, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func readmetadataTag(v metadataTag) assettag.Tag {
	return assettag.Tag{ID: v.ID, Key: v.Key, DisplayName: v.DisplayName, Color: v.Color, LifecycleState: v.LifecycleState, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func projectmetadataType(v customfield.AssetType) metadataType {
	return metadataType{ID: v.ID, Scope: v.Scope, Key: v.Key, DisplayName: v.DisplayName, Description: v.Description, LifecycleState: v.LifecycleState, ExpirationEnabled: v.ExpirationEnabled}
}

func readmetadataType(v metadataType) customfield.AssetType {
	return customfield.AssetType{ID: v.ID, Scope: v.Scope, Key: v.Key, DisplayName: v.DisplayName, Description: v.Description, LifecycleState: v.LifecycleState, ExpirationEnabled: v.ExpirationEnabled}
}

func projectmetadataField(v customfield.Definition) metadataField {
	return metadataField{ID: v.ID, Scope: v.Scope, Key: v.Key, DisplayName: v.DisplayName, Type: v.Type, EnumOptions: v.EnumOptions, Applicability: v.Applicability, CustomAssetTypeIDs: v.CustomAssetTypeIDs, LifecycleState: v.LifecycleState}
}

func readmetadataField(v metadataField) customfield.Definition {
	return customfield.Definition{ID: v.ID, Scope: v.Scope, Key: v.Key, DisplayName: v.DisplayName, Type: v.Type, EnumOptions: v.EnumOptions, Applicability: v.Applicability, CustomAssetTypeIDs: v.CustomAssetTypeIDs, LifecycleState: v.LifecycleState}
}

func projectmetadataAttachment(v media.Attachment) metadataAttachment {
	return metadataAttachment{ID: v.ID, FileName: v.FileName, ContentType: v.ContentType, SHA256: v.SHA256, SizeBytes: v.SizeBytes, CreatedAt: v.CreatedAt, LifecycleState: v.LifecycleState}
}

func readmetadataAttachment(v metadataAttachment) media.Attachment {
	return media.Attachment{ID: v.ID, FileName: v.FileName, ContentType: v.ContentType, SHA256: v.SHA256, SizeBytes: v.SizeBytes, CreatedAt: v.CreatedAt, LifecycleState: v.LifecycleState}
}

func projectmetadataCheckout(v asset.Checkout) metadataCheckout {
	return metadataCheckout{ID: v.ID, State: v.State, CheckedOutAt: v.CheckedOutAt, CheckedOutByPrincipal: v.CheckedOutByPrincipal, CheckoutDetails: v.CheckoutDetails, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func readmetadataCheckout(v metadataCheckout) asset.Checkout {
	return asset.Checkout{ID: v.ID, State: v.State, CheckedOutAt: v.CheckedOutAt, CheckedOutByPrincipal: v.CheckedOutByPrincipal, CheckoutDetails: v.CheckoutDetails, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func projectAsset(v ports.InventoryExportAsset) metadataAsset {
	r := metadataAsset{ID: v.ID, Title: v.Title, Description: v.Description, Kind: v.Kind, ParentAssetID: v.ParentAssetID, CustomAssetTypeID: v.CustomAssetTypeID, LifecycleState: v.LifecycleState, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, ExpirationDate: v.ExpirationDate, ExpirationPrecision: v.ExpirationPrecision, CustomFields: v.CustomFields, TagIDs: v.TagIDs}
	r.Attachments = make([]metadataAttachment, 0, len(v.Attachments))
	for _, m := range v.Attachments {
		r.Attachments = append(r.Attachments, projectmetadataAttachment(m))
	}
	if v.CurrentCheckout != nil {
		c := projectmetadataCheckout(*v.CurrentCheckout)
		r.CurrentCheckout = &c
	}
	return r
}

func readAsset(v metadataAsset) ports.InventoryExportAsset {
	r := ports.InventoryExportAsset{ID: v.ID, Title: v.Title, Description: v.Description, Kind: v.Kind, ParentAssetID: v.ParentAssetID, CustomAssetTypeID: v.CustomAssetTypeID, LifecycleState: v.LifecycleState, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, ExpirationDate: v.ExpirationDate, ExpirationPrecision: v.ExpirationPrecision, CustomFields: v.CustomFields, TagIDs: v.TagIDs}
	r.Attachments = make([]media.Attachment, 0, len(v.Attachments))
	for _, m := range v.Attachments {
		r.Attachments = append(r.Attachments, readmetadataAttachment(m))
	}
	if v.CurrentCheckout != nil {
		c := readmetadataCheckout(*v.CurrentCheckout)
		r.CurrentCheckout = &c
	}
	return r
}
