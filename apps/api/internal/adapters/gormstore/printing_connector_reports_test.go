package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"testing"
	"time"
)

func TestConnectorReportIsAtomicWithAuditAndSurvivesLegacyHeartbeat(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	now := time.Now().UTC()
	saveTenant(t, ctx, s, "tenant", "Home")
	saveInventory(t, ctx, s, "inventory", "tenant", "Home")
	connector := p.Connector{ID: "connector", Scope: p.Scope{TenantID: "tenant", InventoryID: "inventory"}, ServiceAccountID: "worker", State: p.ConnectorActive, CredentialHash: "digest", CredentialVersion: 1, CredentialExpiresAt: now.Add(time.Hour)}
	row := printingConnectorFromDomain(connector)
	if err := s.db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	report := &p.ConnectorReport{Version: "v1", Commit: "commit", Platform: "linux", Architecture: "amd64", Adapters: []p.AdapterCapability{{ID: "brother", ContractVersions: []uint32{1}, Formats: []string{"image/png"}, CompletionEvidence: "physical"}}}
	unavailable := func(p.Connector, bool) (audit.Record, error) { return audit.Record{}, errors.New("audit unavailable") }
	if _, err := s.HeartbeatPrintConnector(ctx, connector, now, unavailable, report); err == nil {
		t.Fatal("changed report committed without audit")
	}
	stored, err := s.FindPrintConnectorCredential(ctx, "digest")
	if err != nil || stored.Report != nil {
		t.Fatal("failed audit leaked report", err)
	}
	changed := func(p.Connector, bool) (audit.Record, error) {
		r := auditRecord(t, "report", "tenant", "inventory", audit.ActionPrintConnectorUpdated)
		r.TargetID = "connector"
		return r, nil
	}
	if _, err = s.HeartbeatPrintConnector(ctx, connector, now, changed, report); err != nil {
		t.Fatal(err)
	}
	s = NewStore(s.db)
	stored, err = s.HeartbeatPrintConnector(ctx, connector, now.Add(time.Minute), unavailable, nil)
	if err != nil || stored.Report == nil || stored.Report.Version != "v1" || !stored.ReportReceivedAt.Equal(now) {
		t.Fatal("legacy heartbeat lost or refreshed capability report", err)
	}
	stored, err = s.HeartbeatPrintConnector(ctx, connector, now.Add(2*time.Minute), unavailable, report)
	if err != nil || !stored.ReportReceivedAt.Equal(now.Add(2*time.Minute)) {
		t.Fatal("unchanged report required duplicate audit", err)
	}
}
