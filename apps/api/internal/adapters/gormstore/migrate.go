package gormstore

import (
	"context"
	"gorm.io/gorm"
)

func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&userModel{}, &tenantModel{}, &inventoryModel{}, &inventoryAccessGrantModel{}, &inventoryAccessInvitationModel{}, &customAssetTypeModel{}, &customFieldDefinitionModel{}, &customFieldDefinitionAssetTypeModel{}, &assetModel{}, &assetCheckoutModel{}, &assetTagModel{}, &assetTagAssignmentModel{}, &undoableOperationModel{}, &attachmentModel{}, &blobDeletionEventModel{}, &auditRecordModel{}, &authorizationOutboxEventModel{}, &providerProfileModel{}, &providerCredentialModel{}, &voiceProviderConfigurationModel{}, &conversationWorkflowModel{}, &conversationWorkflowRevisionModel{}, &conversationWorkflowSelectionModel{}, &evaluationCaseModel{}, &evaluationCaseRevisionModel{}, &realtimeSessionModel{}, &actionPlanModel{}, &importJobModel{}, &importJobSourceModel{}, &importSourceLinkModel{}, &importJobResourceModel{})
}
