package gormstore

import (
	"context"
	"gorm.io/gorm"
)

func Migrate(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).AutoMigrate(&userModel{}, &tenantModel{}, &inventoryModel{}, &inventoryAccessGrantModel{}, &inventoryAccessInvitationModel{}, &customAssetTypeModel{}, &customFieldDefinitionModel{}, &customFieldDefinitionAssetTypeModel{}, &assetModel{}, &assetCheckoutModel{}, &assetTagModel{}, &assetTagAssignmentModel{}, &undoableOperationModel{}, &attachmentModel{}, &mediaBlobKeyModel{}, &thumbnailJobModel{}, &thumbnailBackfillModel{}, &blobDeletionEventModel{}, &auditRecordModel{}, &authorizationOutboxEventModel{}, &providerProfileModel{}, &providerCredentialModel{}, &voiceProviderConfigurationModel{}, &conversationWorkflowModel{}, &conversationWorkflowRevisionModel{}, &conversationWorkflowSelectionModel{}, &evaluationRunModel{}, &evaluationCaseModel{}, &evaluationCaseRevisionModel{}, &realtimeSessionModel{}, &actionPlanModel{}, &importJobModel{}, &importJobSourceModel{}, &importSourceLinkModel{}, &importJobResourceModel{}); err != nil {
		return err
	}
	return seedMediaBlobKeys(db.WithContext(ctx))
}
