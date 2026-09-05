package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/testsupport/workflowdiscovery"
	"testing"
)

func TestWorkflowDiscoveryRepository(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, "discovery-home", "Home")
	workflowdiscovery.Verify(t, store)
}
