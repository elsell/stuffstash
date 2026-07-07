package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestNormalizedImportPlanForJobDoesNotMutateAttachmentContent(t *testing.T) {
	application := New(Dependencies{})
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Attachments[0].Content = pngAttachmentBytes()

	normalized, err := application.normalizedImportPlanForJob(context.Background(), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), plan)
	if err != nil {
		t.Fatalf("normalize plan: %v", err)
	}
	if len(normalized.Attachments[0].Content) != 0 {
		t.Fatalf("expected normalized plan to strip attachment bytes")
	}
	if len(plan.Attachments[0].Content) == 0 {
		t.Fatalf("expected original plan attachment bytes to remain available for apply")
	}
}

func TestSourceFingerprintIgnoresAttachmentOrder(t *testing.T) {
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Attachments = []importplan.Attachment{
		{SourceID: "attachment:two", AssetSourceID: "source:drill", FileName: "two.jpg", ContentType: "image/jpeg", Primary: false},
		{SourceID: "attachment:one", AssetSourceID: "source:drill", FileName: "one.jpg", ContentType: "image/jpeg", Primary: true},
	}
	swapped := plan
	swapped.Attachments = []importplan.Attachment{plan.Attachments[1], plan.Attachments[0]}

	first, err := sourceFingerprint(plan)
	if err != nil {
		t.Fatalf("first fingerprint: %v", err)
	}
	second, err := sourceFingerprint(swapped)
	if err != nil {
		t.Fatalf("second fingerprint: %v", err)
	}
	if first != second {
		t.Fatalf("expected attachment order-insensitive fingerprint, got %q and %q", first, second)
	}
}

func TestSourceFingerprintIgnoresAttachmentByteMetadataButNotIdentity(t *testing.T) {
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Attachments = []importplan.Attachment{{
		SourceID:      "attachment:one",
		AssetSourceID: "source:drill",
		FileName:      "preview-name.jpg",
		ContentType:   "image/jpeg",
		SizeBytes:     10,
		Primary:       true,
		Content:       []byte("preview bytes"),
	}}
	applyPlan := plan
	applyPlan.Attachments = []importplan.Attachment{{
		SourceID:      "attachment:one",
		AssetSourceID: "source:drill",
		FileName:      "sniffed-safe-name.png",
		ContentType:   "image/png",
		SizeBytes:     4096,
		Primary:       true,
		Content:       pngAttachmentBytes(),
	}}

	previewFingerprint, err := sourceFingerprint(plan)
	if err != nil {
		t.Fatalf("preview fingerprint: %v", err)
	}
	applyFingerprint, err := sourceFingerprint(applyPlan)
	if err != nil {
		t.Fatalf("apply fingerprint: %v", err)
	}
	if previewFingerprint != applyFingerprint {
		t.Fatalf("expected fingerprint to ignore attachment bytes and downloaded metadata, got %q and %q", previewFingerprint, applyFingerprint)
	}

	changedIdentity := applyPlan
	changedIdentity.Attachments = []importplan.Attachment{{
		SourceID:      "attachment:two",
		AssetSourceID: "source:drill",
		Primary:       true,
	}}
	changedPrimary := applyPlan
	changedPrimary.Attachments = []importplan.Attachment{{
		SourceID:      "attachment:one",
		AssetSourceID: "source:drill",
		Primary:       false,
	}}
	for name, candidate := range map[string]importplan.Plan{
		"attachment identity": changedIdentity,
		"primary status":      changedPrimary,
	} {
		fingerprint, err := sourceFingerprint(candidate)
		if err != nil {
			t.Fatalf("%s fingerprint: %v", name, err)
		}
		if fingerprint == previewFingerprint {
			t.Fatalf("expected %s to remain a fingerprint input", name)
		}
	}
}
