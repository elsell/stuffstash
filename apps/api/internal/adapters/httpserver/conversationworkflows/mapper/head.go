package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/dto"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func HeadToResponse(value ports.WorkflowHeadRecord) dto.WorkflowHead {
	return dto.WorkflowHead{ID: string(value.ID), Name: value.Name, LatestRevisionID: string(value.LatestRevisionID), LatestRevision: value.LatestRevision, ActiveRevisionID: string(value.ActiveRevisionID), CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
func SelectionToResponse(value *ports.WorkflowSelectionReference) *dto.WorkflowSelection {
	if value == nil {
		return nil
	}
	return &dto.WorkflowSelection{WorkflowID: string(value.WorkflowID), RevisionID: string(value.RevisionID)}
}
