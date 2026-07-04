import type { AuditScope } from '$lib/domain/inventory';

export type AuditStatusKind = 'none' | 'missing-context' | 'denied' | 'error' | 'loading' | 'empty';

export interface AuditStatusPresentation {
  kind: AuditStatusKind;
  message: string;
  role?: 'alert' | 'status';
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
