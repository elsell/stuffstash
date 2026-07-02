package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestValidateActionPlanPhotoAttachmentMetadataAcceptsAttachableProposedCommand(t *testing.T) {
	t.Parallel()

	application := newActionPlanTestApp(&fakeActionPlanRepository{records: map[string]ports.ActionPlanRecord{
		"plan-1": actionPlanRecord("plan-1", actionplan.StateProposed),
	}}, &fakeIDGenerator{}, nil)

	err := application.ValidateActionPlanPhotoAttachmentMetadata(context.Background(), ActionPlanPhotoAttachmentMetadataInput{
		Decision: actionPlanPhotoDecision(),
		Photos: []ActionPlanPhotoAttachmentMetadata{{
			CommandID:   "command-1",
			FileName:    "water-bottle.jpg",
			ContentType: "image/jpeg",
			SizeBytes:   128,
		}},
	})
	if err != nil {
		t.Fatalf("validate photo attachment metadata: %v", err)
	}
}

func TestValidateActionPlanPhotoAttachmentMetadataRejectsUnsafeMetadataBeforeApproval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		record  ports.ActionPlanRecord
		photo   ActionPlanPhotoAttachmentMetadata
		wantErr error
	}{
		{
			name:   "unsupported content type",
			record: actionPlanRecord("plan-1", actionplan.StateProposed),
			photo: ActionPlanPhotoAttachmentMetadata{
				CommandID:   "command-1",
				FileName:    "water-bottle.pdf",
				ContentType: "application/pdf",
				SizeBytes:   128,
			},
			wantErr: ErrInvalidInput,
		},
		{
			name:   "oversized photo",
			record: actionPlanRecord("plan-1", actionplan.StateProposed),
			photo: ActionPlanPhotoAttachmentMetadata{
				CommandID:   "command-1",
				FileName:    "water-bottle.jpg",
				ContentType: "image/jpeg",
				SizeBytes:   26 * 1024 * 1024,
			},
			wantErr: ErrInvalidInput,
		},
		{
			name:   "unknown command",
			record: actionPlanRecord("plan-1", actionplan.StateProposed),
			photo: ActionPlanPhotoAttachmentMetadata{
				CommandID:   "command-missing",
				FileName:    "water-bottle.jpg",
				ContentType: "image/jpeg",
				SizeBytes:   128,
			},
			wantErr: ErrInvalidInput,
		},
		{
			name:   "non attachable command",
			record: actionPlanRecordWithCommand("plan-1", actionplan.StateProposed, actionplan.CommandKindArchiveAsset, `{"assetId":"asset-1"}`),
			photo: ActionPlanPhotoAttachmentMetadata{
				CommandID:   "command-1",
				FileName:    "water-bottle.jpg",
				ContentType: "image/jpeg",
				SizeBytes:   128,
			},
			wantErr: ErrInvalidInput,
		},
		{
			name:   "already approved plan",
			record: actionPlanRecord("plan-1", actionplan.StateApproved),
			photo: ActionPlanPhotoAttachmentMetadata{
				CommandID:   "command-1",
				FileName:    "water-bottle.jpg",
				ContentType: "image/jpeg",
				SizeBytes:   128,
			},
			wantErr: ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			application := newActionPlanTestApp(&fakeActionPlanRepository{records: map[string]ports.ActionPlanRecord{
				"plan-1": tt.record,
			}}, &fakeIDGenerator{}, nil)

			err := application.ValidateActionPlanPhotoAttachmentMetadata(context.Background(), ActionPlanPhotoAttachmentMetadataInput{
				Decision: actionPlanPhotoDecision(),
				Photos:   []ActionPlanPhotoAttachmentMetadata{tt.photo},
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func actionPlanPhotoDecision() ActionPlanDecisionInput {
	return ActionPlanDecisionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		PlanID:      "plan-1",
	}
}
