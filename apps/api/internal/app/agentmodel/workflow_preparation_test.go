package agentmodel

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestWorkflowPreparationRequiresAccessAndFailsClosedOnBrokenSelection(t *testing.T) {
	repository := newWorkflowFakeRepository()
	repository.selected = map[tenant.ID]ports.WorkflowSelectionReference{"tenant-home": {WorkflowID: "missing", RevisionID: "missing"}}
	service := NewConversationWorkflowService(ConversationWorkflowDependencies{Authorizer: denyTenantAuthorizer{}, Repository: repository, Limits: workflowServiceLimits()})
	input := PrepareWorkflowInput{Principal: testPrincipal(), TenantID: "tenant-home"}
	if _, err := service.PrepareSelected(context.Background(), input); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("denial lost: %v", err)
	}
	service.deps.Authorizer = workflowViewAuthorizer{}
	if _, err := service.PrepareSelected(context.Background(), input); !errors.Is(err, apperrors.ErrPrecondition) {
		t.Fatalf("broken selection silently defaulted: %v", err)
	}
	delete(repository.selected, "tenant-home")
	if prepared, err := service.PrepareSelected(context.Background(), input); err != nil || prepared != nil {
		t.Fatalf("absent selection should use default: %v", err)
	}
}

type workflowViewAuthorizer struct{ allowTenantConfigureAuthorizer }

func (workflowViewAuthorizer) CheckTenant(_ context.Context, _ identity.Principal, permission ports.TenantPermission, _ tenant.ID) error {
	if permission == ports.TenantPermissionView || permission == ports.TenantPermissionConfigure {
		return nil
	}
	return ports.ErrForbidden
}
