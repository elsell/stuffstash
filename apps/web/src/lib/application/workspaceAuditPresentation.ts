import { timestampLabel } from '$lib/presentation/timestamp';
import { auditVocabularyLabel } from './workspaceAuditVocabulary';
import { t } from '$lib/presentation/localization';
import type { AuditRecord, AuditScope } from '$lib/domain/inventory';

export type AuditStatusKind = 'none' | 'missing-context' | 'denied' | 'error' | 'loading' | 'empty';

export interface AuditStatusPresentation {
  kind: AuditStatusKind;
  message: string;
  role?: 'alert' | 'status';
}

export interface AuditTechnicalDetail {
  label: string;
  value: string;
}

export interface AuditRecordPresentation {
  title: string;
  actorLabel: string;
  sourceLabel: string;
  targetLabel: string;
  occurredAtLabel: string;
  primaryText: string;
  technicalDetails: AuditTechnicalDetail[];
}

export interface AuditDayGroup {
  key: string;
  label: string;
  records: AuditRecord[];
}

export function groupAuditRecordsByDay(
  records: AuditRecord[],
  locale?: string,
  timeZone?: string
): AuditDayGroup[] {
  const formatter = new Intl.DateTimeFormat(locale, { dateStyle: 'long', timeZone });
  const groups = new Map<string, AuditDayGroup>();
  for (const record of records) {
    const date = new Date(record.occurredAt);
    const valid = !Number.isNaN(date.getTime());
    const key = valid
      ? new Intl.DateTimeFormat('en-CA', { year: 'numeric', month: '2-digit', day: '2-digit', timeZone }).format(date)
      : record.occurredAt;
    const label = valid ? formatter.format(date) : t('web.workspaceAuditPresentation.dateUnavailable');
    const group = groups.get(key) ?? { key, label, records: [] };
    group.records.push(record);
    groups.set(key, group);
  }
  return [...groups.values()];
}

export function auditStatusPresentation(input: {
  hasTenant: boolean;
  hasInventory: boolean;
  scope: AuditScope;
  canReadScope: boolean;
  error: string;
  busy: boolean;
  loaded: boolean;
  recordCount: number;
}): AuditStatusPresentation {
  if (!input.hasTenant || (input.scope === 'inventory' && !input.hasInventory)) {
    return { kind: 'missing-context', message: t('web.workspaceAuditPresentation.selectAnInventoryBeforeViewingAuditHistory') };
  }
  if (!input.canReadScope) {
    return {
      kind: 'denied',
      message:
        input.scope === 'tenant'
          ? t('web.workspaceAuditPresentation.tenantAuditHistoryRequiresTenantConfigurationAccess')
          : t('web.workspaceAuditPresentation.inventoryAuditHistoryRequiresInventoryViewAccess'),
      role: 'alert'
    };
  }
  if (input.error) {
    return { kind: 'error', message: input.error, role: 'alert' };
  }
  if (input.busy && !input.loaded) {
    return { kind: 'loading', message: t('web.workspaceAuditPresentation.loadingAuditHistory'), role: 'status' };
  }
  if (input.loaded && input.recordCount === 0) {
    return { kind: 'empty', message: t('web.workspaceAuditPresentation.noAuditRecordsFound') };
  }
  return { kind: 'none', message: '' };
}

export function auditRecordPresentation(record: AuditRecord): AuditRecordPresentation {
  const title = humanizeAction(record.action);
  const actorLabel = humanizePrincipal(record.principalId);
  const sourceLabel = humanizeSource(record.source);
  const targetLabel = humanizeTarget(record.targetType);
  const occurredAtLabel = humanizeDate(record.occurredAt);
  return {
    title,
    actorLabel,
    sourceLabel,
    targetLabel,
    occurredAtLabel,
    primaryText: `${title} ${actorLabel} ${sourceLabel} ${targetLabel} ${occurredAtLabel}`,
    technicalDetails: [
      { label: t('web.workspaceAuditPresentation.actionCode'), value: record.action },
      { label: t('audit.targetType'), value: record.targetType },
      { label: t('web.workspaceAuditPresentation.targetID'), value: record.targetId },
      { label: t('web.workspaceAuditPresentation.principalID'), value: record.principalId },
      { label: t('web.workspaceAuditPresentation.source'), value: record.source },
      ...(record.requestId ? [{ label: t('web.workspaceAuditPresentation.requestID'), value: record.requestId }] : []),
      ...Object.entries(record.metadata).map(([key, value]) => ({ label: t('web.workspaceAuditPresentation.metadata', { value: String(humanizeMetadataKey(key)) }), value }))
    ].filter((detail) => detail.value.trim().length > 0)
  };
}

function humanizeAction(value: string): string {
  return auditVocabularyLabel('action', value) ?? t('web.workspaceAuditPresentation.activityRecorded');
}

function humanizePrincipal(value: string): string {
  if (!value.trim()) {
    return t('web.workspaceAuditPresentation.unknownActor');
  }
  if (value.includes('@')) {
    return value;
  }
  if (value === 'api') {
    return t('web.workspaceAuditPresentation.api');
  }
  if (value === 'principal-owner') {
    return t('web.workspaceAuditPresentation.owner');
  }
  if (value.startsWith('oidc_') || value.startsWith('oidc:')) {
    return t('web.workspaceAuditPresentation.signedInUser');
  }
  if (value.startsWith('principal-')) {
    return t('web.workspaceAuditPresentation.user');
  }
  if (value === 'system') {
    return t('web.workspaceAuditPresentation.system');
  }
  return t('web.workspaceAuditPresentation.user');
}

function humanizeSource(value: string): string {
  const knownSources: Record<string, string> = {
    api: t('web.workspaceAuditPresentation.api'),
    web: t('web.workspaceAuditPresentation.web'),
    mobile: t('web.workspaceAuditPresentation.mobile'),
    system: t('web.workspaceAuditPresentation.system'),
    import: t('web.workspaceAuditPresentation.import'),
    local_demo: t('web.workspaceAuditPresentation.localDemo')
  };
  return auditVocabularyLabel('source', value) ?? (Object.hasOwn(knownSources, value) ? knownSources[value] : undefined) ?? t('web.workspaceAuditPresentation.recordedSource');
}

function humanizeTarget(value: string): string {
  const knownTargets: Record<string, string> = {
    asset: t('web.workspaceAuditPresentation.asset'),
    inventory: t('web.workspaceAuditPresentation.inventory'),
    tenant: t('web.workspaceAuditPresentation.tenant'),
    attachment: t('web.workspaceAuditPresentation.attachment'),
    invitation: t('web.workspaceAuditPresentation.invitation'),
    custom_field: t('web.workspaceAuditPresentation.customField'),
    custom_asset_type: t('web.workspaceAuditPresentation.customAssetType')
  };
  return auditVocabularyLabel('target', value) ?? (Object.hasOwn(knownTargets, value) ? knownTargets[value] : undefined) ?? t('audit.target.unknown');
}

function humanizeMetadataKey(value: string): string {
  return value.replaceAll('_', ' ');
}

function humanizeDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return timestampLabel(value);
}
