import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { t } from '$lib/presentation/localization';
import { auditRecordPresentation } from './workspaceAuditPresentation';

const domain = readFileSync('../api/internal/domain/audit/audit.go', 'utf8');
const vocabulary = (type: string) => [...domain.matchAll(new RegExp(`\\s${type} = "([^"]+)"`, 'g'))].map(match => match[1]);
const record = { id: 'audit', tenantId: 'tenant', inventoryId: 'inventory', principalId: 'system', action: 'asset.moved', source: 'api', targetType: 'asset', targetId: 'asset', occurredAt: '2026-10-01T12:00:00Z', metadata: {} };

describe('supported audit vocabulary', () => {
  it('presents every supported action, target and source through a catalog message', () => {
    for (const [type, field, label, fallback] of [
      ['Action', 'action', 'title', 'web.workspaceAuditPresentation.activityRecorded'],
      ['TargetType', 'targetType', 'targetLabel', 'audit.target.unknown'],
      ['Source', 'source', 'sourceLabel', 'web.workspaceAuditPresentation.recordedSource']
    ] as const) {
      const values = vocabulary(type);
      expect(values.length).toBeGreaterThan(0);
      for (const code of values) {
        const result = auditRecordPresentation({ ...record, [field]: code });
        expect(result[label], code).not.toBe(t(fallback));
        if (import.meta.env.VITE_STUFF_STASH_UI_LOCALE === 'en-XA') expect(result[label], code).toMatch(/^\[.*~\]$/);
      }
    }
  });
});
