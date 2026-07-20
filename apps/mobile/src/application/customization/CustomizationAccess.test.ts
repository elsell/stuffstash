import { describe, expect, it } from 'vitest';
import { CustomizationAccessPolicy } from './CustomizationAccess';
import { BufferedCustomizationObservability } from './CustomizationObservability';
import type { CustomizationContext } from './CustomizationRepository';

const context: CustomizationContext = {
  tenantId: 'tenant-one', tenantName: 'Home', tenantPermissions: [],
  inventoryId: 'inventory-one', inventoryName: 'Household', inventoryPermissions: ['view']
};

describe('CustomizationAccessPolicy', () => {
  it('allows effective inventory inspection while denying local mutation and inherited mutation', () => {
    const events = new BufferedCustomizationObservability();
    const policy = new CustomizationAccessPolicy(events);
    expect(policy.readOrRecord(context, 'field', 'inventory')).toBe(true);
    expect(policy.mutationOrRecord(context, 'field', 'inventory')).toBe(false);
    expect(policy.mutationOrRecord({ ...context, inventoryPermissions: ['view', 'configure'] }, 'field', 'inventory', true)).toBe(false);
    expect(events.events()).toHaveLength(2);
  });

  it('fails closed before tenant-only reads without tenant configure', () => {
    const events = new BufferedCustomizationObservability();
    const policy = new CustomizationAccessPolicy(events);
    expect(policy.readOrRecord(context, 'asset-type', 'tenant')).toBe(false);
    expect(events.events()[0]).toEqual({ name: 'customization.permission_denied', resource: 'asset-type', scope: 'tenant' });
  });
});
