package assets

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

type auditRecordInput = appsupport.AuditRecordInput

func (s Service) newAuditRecord(input auditRecordInput) (audit.Record, error) {
	return appsupport.NewAuditRecord(s.ids, s.clock, input)
}

func (s Service) saveReadAuditRecord(ctx context.Context, input auditRecordInput) error {
	return appsupport.SaveReadAuditRecord(ctx, s.audit, s.ids, s.clock, input)
}
