package dataportability

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type archiveSequence struct{ value int }

func (s *archiveSequence) NewID() string { s.value++; return "new-" + strconv.Itoa(s.value) }
func TestRestorePlanRemapsWholeGraphAndRespectsSelection(t *testing.T) {
	source := validArchiveDocument()
	now := source.ExportedAt
	source.Assets[1].CurrentCheckout = &asset.Checkout{ID: "checkout", State: asset.CheckoutStateOpen, CheckedOutAt: now, CheckedOutByPrincipal: "old-owner", CheckoutDetails: "In use", CreatedAt: now, UpdatedAt: now}
	source.Assets[1].Attachments = []media.Attachment{{ID: "photo", ContentType: media.ContentTypeJPEG, FileName: "photo.jpg", SHA256: media.SHA256(strings.Repeat("a", 64)), SizeBytes: 10, CreatedAt: now, LifecycleState: media.LifecycleStateActive}, {ID: "file", ContentType: media.ContentTypePDF, FileName: "receipt.pdf", SHA256: media.SHA256(strings.Repeat("b", 64)), SizeBytes: 20, CreatedAt: now, LifecycleState: media.LifecycleStateActive}}
	dest := ports.ArchiveRestoreDestination{TenantID: "new-tenant", InventoryID: "new-inventory", Name: "Restored home", PrincipalID: "new-owner"}
	plan, err := BuildArchiveRestorePlan(context.Background(), source, dest, ports.ArchiveKeyReservations{Fields: map[string]bool{"strength": true}, Types: map[string]bool{"medicine": true}}, ports.ArchiveMediaSelection{Photos: true}, &archiveSequence{}, 100)
	if err != nil {
		t.Fatal(err)
	}
	d := plan.Document
	a := d.Assets[1]
	if d.TenantID != dest.TenantID || d.InventoryID != dest.InventoryID || d.InventoryName != dest.Name || a.ID == source.Assets[1].ID || a.ParentAssetID != d.Assets[0].ID || a.CustomAssetTypeID != d.CustomAssetTypes[0].ID.String() || a.TagIDs[0] != d.Tags[0].ID.String() || d.CustomFieldDefinitions[0].CustomAssetTypeIDs[0] != d.CustomAssetTypes[0].ID {
		t.Fatal("reference graph not remapped")
	}
	if d.CustomAssetTypes[0].Scope != "inventory" || d.CustomFieldDefinitions[0].Scope != "inventory" || len(plan.KeyRemappings) != 2 || a.CustomFields[d.CustomFieldDefinitions[0].Key.String()] != float64(10) {
		t.Fatal("inherited/colliding definitions lost")
	}
	if plan.OmittedAttachments != 1 || len(a.Attachments) != 1 || a.Attachments[0].ID == "photo" || a.Attachments[0].StorageKey.String() != dest.TenantID+"/"+dest.InventoryID+"/"+a.ID+"/"+a.Attachments[0].ID.String() {
		t.Fatal("media selection or destination key incorrect")
	}
	if a.CurrentCheckout.CheckedOutByPrincipal != "new-owner" || plan.CheckoutSourcePrincipals[a.CurrentCheckout.ID.String()] != "old-owner" || a.CurrentCheckout.ID == "checkout" {
		t.Fatal("checkout actor not remapped")
	}
	if !a.CreatedAt.Equal(source.Assets[1].CreatedAt) || a.LifecycleState != "archived" || a.CurrentCheckout.CheckoutDetails != "In use" {
		t.Fatal("source metadata changed")
	}
	if source.Assets[1].TagIDs[0] != "tag" || source.Assets[1].CurrentCheckout.CheckedOutByPrincipal != "old-owner" || source.CustomFieldDefinitions[0].CustomAssetTypeIDs[0] != "type" {
		t.Fatal("source document mutated")
	}
}
