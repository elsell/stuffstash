package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/testsupport/workflowdiscovery"
	"testing"
)

func TestWorkflowDiscoveryRepository(t *testing.T) {
	store := NewStore()
	name, _ := tenant.NewName("Home")
	if err := store.SaveTenant(context.Background(), tenant.Tenant{ID: "discovery-home", Name: name}); err != nil {
		t.Fatal(err)
	}
	workflowdiscovery.Verify(t, store)
}
