package worklimit

import "testing"

func TestFlightsSerializeOnlyTheSameOriginalAndForgetReleasedOwners(t *testing.T) {
	f := NewThumbnailFlights()
	release, _, owned := f.TryStart("tenant/inventory/photo")
	if !owned {
		t.Fatal("first owner rejected")
	}
	_, done, owned := f.TryStart("tenant/inventory/photo")
	if owned {
		t.Fatal("same photo admitted twice")
	}
	other, _, owned := f.TryStart("other-tenant/inventory/photo")
	if !owned {
		t.Fatal("unrelated photo blocked")
	}
	other()
	release()
	select {
	case <-done:
	default:
		t.Fatal("waiter was not woken")
	}
	fresh, freshDone, owned := f.TryStart("tenant/inventory/photo")
	if !owned {
		t.Fatal("released key retained")
	}
	release()
	select {
	case <-freshDone:
		t.Fatal("old release closed new flight")
	default:
	}
	fresh()
}
