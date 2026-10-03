package printing

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func (s *JobService) CleanupInterval() time.Duration { return s.config.CleanupInterval }
func (s *JobService) Maintain(ctx context.Context, after string) (ports.PrintMaintenancePage, error) {
	now := s.labels.deps.Clock.Now()
	return s.jobs.MaintainPrintJobs(ctx, ports.PrintJobMaintenance{Now: now, TerminalBefore: now.Add(-s.config.TerminalTTL), Limit: 100, After: after, Audit: func(before, after p.Job) (audit.Record, error) {
		action := audit.ActionPrintJobFailed
		if after.Status == p.JobQueued {
			action = audit.ActionPrintJobReleased
		}
		if after.Status == p.JobUncertain {
			action = audit.ActionPrintJobUncertain
		}
		return appsupport.NewAuditRecord(s.labels.deps.IDs, s.labels.deps.Clock, appsupport.AuditRecordInput{Principal: identity.Principal{ID: identity.PrincipalID(after.RequestedBy)}, TenantID: tenant.ID(after.Scope.TenantID), InventoryID: inventory.InventoryID(after.Scope.InventoryID), Source: audit.SourceBackgroundJob, Action: action, TargetType: audit.TargetPrintJob, TargetID: string(after.ID)})
	}})
}
