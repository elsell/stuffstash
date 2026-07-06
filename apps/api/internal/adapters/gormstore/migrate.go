package gormstore

import (
	"context"
	"gorm.io/gorm"
)

func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&userModel{}, &tenantModel{}, &inventoryModel{}, &inventoryAccessGrantModel{}, &inventoryAccessInvitationModel{}, &customAssetTypeModel{}, &customFieldDefinitionModel{}, &customFieldDefinitionAssetTypeModel{}, &assetModel{}, &undoableOperationModel{}, &attachmentModel{}, &blobDeletionEventModel{}, &auditRecordModel{}, &authorizationOutboxEventModel{}, &providerProfileModel{}, &providerCredentialModel{}, &voiceProviderConfigurationModel{}, &realtimeSessionModel{}, &actionPlanModel{}, &importJobModel{}, &importJobSourceModel{}, &importSourceLinkModel{}, &importJobResourceModel{})
}
