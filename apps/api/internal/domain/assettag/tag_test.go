package assettag

import (
	"testing"
	"time"
)

func TestTagValueObjectsValidateAndNormalize(t *testing.T) {
	key, ok := NewKey("garage-tools")
	if !ok || key.String() != "garage-tools" {
		t.Fatalf("expected key to validate, got %q ok=%v", key, ok)
	}
	if _, ok := NewKey("Garage Tools"); ok {
		t.Fatal("expected display-style key to be rejected")
	}
	fromName, ok := KeyFromDisplayName("Garage Tools!")
	if !ok || fromName.String() != "garage-tools" {
		t.Fatalf("expected key from display name, got %q ok=%v", fromName, ok)
	}
	color, ok := NewColor("2f80ed")
	if !ok || color.String() != "#2F80ED" {
		t.Fatalf("expected color normalization, got %q ok=%v", color, ok)
	}
	if _, ok := NewColor("#12zz99"); ok {
		t.Fatal("expected invalid color rejection")
	}
}

func TestNewTagRequiresInventoryScope(t *testing.T) {
	key, _ := NewKey("tools")
	name, _ := NewDisplayName("Tools")
	color, _ := NewColor("#2f80ed")
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)

	tag, ok := NewTag("tag-one", "tenant-one", "inventory-one", key, name, color, now)
	if !ok {
		t.Fatal("expected valid tag")
	}
	if tag.LifecycleState != LifecycleStateActive || tag.Color.String() != "#2F80ED" {
		t.Fatalf("unexpected tag = %#v", tag)
	}

	if _, ok := NewTag("", "tenant-one", "inventory-one", key, name, color, now); ok {
		t.Fatal("expected missing id to be invalid")
	}
}
