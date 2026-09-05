package memory

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

func TestEvaluationRunRepository(t *testing.T) {
	store := NewStore()
	name, _ := tenant.NewName("Home")
	if err := store.SaveTenant(context.Background(), tenant.Tenant{ID: evaluationrun.TenantID, Name: name}); err != nil {
		t.Fatal(err)
	}
	evaluationrun.Verify(t, store)
	evaluationrun.VerifyConcurrentClaims(t, store)
}
