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
    return { kind: 'missing-context', message: 'Select an inventory before viewing audit history.' };
  }
  if (!input.canReadScope) {
    return {
      kind: 'denied',
      message:
        input.scope === 'tenant'
          ? 'Tenant audit history requires tenant configuration access.'
          : 'Inventory audit history requires inventory view access.',
      role: 'alert'
    };
  }
  if (input.error) {
    return { kind: 'error', message: input.error, role: 'alert' };
  }
  if (input.busy && !input.loaded) {
    return { kind: 'loading', message: 'Loading audit history...', role: 'status' };
  }
  if (input.loaded && input.recordCount === 0) {
    return { kind: 'empty', message: 'No audit records found.' };
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
      { label: 'Action code', value: record.action },
      { label: 'Target ID', value: record.targetId },
      { label: 'Principal ID', value: record.principalId },
      { label: 'Source', value: record.source },
      ...(record.requestId ? [{ label: 'Request ID', value: record.requestId }] : []),
      ...Object.entries(record.metadata).map(([key, value]) => ({ label: `Metadata ${humanizeMetadataKey(key)}`, value }))
    ].filter((detail) => detail.value.trim().length > 0)
  };
}

function humanizeAction(value: string): string {
  const knownActions: Record<string, string> = {
    'asset.created': 'Asset created',
    'asset.updated': 'Asset updated',
    'asset.archived': 'Asset archived',
    'asset.restored': 'Asset restored',
    'asset.deleted': 'Asset deleted',
    'attachment.created': 'Attachment added',
    'attachment.deleted': 'Attachment removed',
    'inventory.created': 'Inventory created',
    'tenant.created': 'Tenant created'
  };
  return knownActions[value] ?? sentenceCase(value);
}

function humanizePrincipal(value: string): string {
  if (!value.trim()) {
    return 'Unknown actor';
  }
  if (value.includes('@')) {
    return value;
  }
  if (value === 'api') {
    return 'API';
  }
  if (value === 'principal-owner') {
    return 'Owner';
  }
  if (value.startsWith('oidc_') || value.startsWith('oidc:')) {
    return 'Signed-in user';
  }
  if (value.startsWith('principal-')) {
    return 'User';
  }
  if (value === 'system') {
    return 'System';
  }
  return 'User';
}

function humanizeSource(value: string): string {
  const knownSources: Record<string, string> = {
    api: 'API',
    web: 'Web',
    mobile: 'Mobile',
    system: 'System',
    import: 'Import',
    local_demo: 'Local demo'
  };
  return knownSources[value] ?? 'Recorded source';
}

function humanizeTarget(value: string): string {
  const knownTargets: Record<string, string> = {
    asset: 'Asset',
    inventory: 'Inventory',
    tenant: 'Tenant',
    attachment: 'Attachment',
    invitation: 'Invitation',
    custom_field: 'Custom field',
    custom_asset_type: 'Custom asset type'
  };
  return knownTargets[value] ?? sentenceCase(value);
}

function humanizeMetadataKey(value: string): string {
  return value.replaceAll('_', ' ');
}

function humanizeDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function sentenceCase(value: string): string {
  const words = value
    .replaceAll('.', ' ')
    .replaceAll('_', ' ')
    .replaceAll('-', ' ')
    .trim()
    .replace(/\s+/g, ' ');
  if (!words) {
    return 'Activity recorded';
  }
  return words.charAt(0).toUpperCase() + words.slice(1);
}
