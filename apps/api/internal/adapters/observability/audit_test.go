package observability

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestAuditObservationPreservesRecordsScopeAndConflicts(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	name, _ := tenant.NewName("Private home")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: name}); err != nil {
		t.Fatal(err)
	}
	record, ok := audit.NewRecord("audit-one", "home", "", "principal", audit.ActionTenantCreated, audit.SourceAPI, audit.TargetTenant, "home", time.Now(), "private-request", map[string]string{"private": "private-value"})
	if !ok {
		t.Fatal("invalid test record")
	}
	measurements := &recordingOperations{}
	repository := ObserveAudit(store, measurements)
	if err := repository.SaveAuditRecord(ctx, record); err != nil {
		t.Fatal(err)
	}
	if err := repository.SaveAuditRecord(ctx, record); !errors.Is(err, ports.ErrConflict) {
		t.Fatal("audit conflict changed")
	}
	records, err := repository.ListTenantAuditRecords(ctx, "home", ports.AuditRecordPageRequest{Limit: 10})
	if err != nil || len(records) != 1 || records[0].ID != record.ID || records[0].Metadata["private"] != "private-value" {
		t.Fatal("audit record changed")
	}
	other, err := repository.ListTenantAuditRecords(ctx, "other", ports.AuditRecordPageRequest{Limit: 10})
	if err != nil || len(other) != 0 {
		t.Fatal("audit scope changed")
	}
	if len(measurements.operations) != 4 || measurements.operations[0] != ports.OperationAuditWrite || measurements.operations[2] != ports.OperationAuditRead || !errors.Is(measurements.results[1], ports.ErrConflict) {
		t.Fatal("audit measurements missing")
	}
}
