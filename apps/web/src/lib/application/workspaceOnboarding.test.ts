import { describe, expect, it } from 'vitest';
import {
  validateWorkspaceSetupDraft,
  workspaceSetupDescription,
  workspaceSetupTitle
} from './workspaceOnboarding';

describe('workspace onboarding helpers', () => {
  it('requires tenant and inventory names when creating the first workspace', () => {
    expect(validateWorkspaceSetupDraft('tenant_and_inventory', { tenantName: ' ', inventoryName: '' })).toEqual({
      valid: false,
      tenantName: '',
      inventoryName: '',
      tenantError: 'Name your tenant.',
      inventoryError: 'Name your inventory.'
    });
  });

  it('requires only an inventory name when a tenant already exists', () => {
    expect(validateWorkspaceSetupDraft('inventory', { tenantName: '', inventoryName: ' Garage ' })).toEqual({
      valid: true,
      tenantName: '',
      inventoryName: 'Garage',
      tenantError: '',
      inventoryError: ''
    });
  });

  it('describes setup mode without inventing default names', () => {
    expect(workspaceSetupTitle('tenant_and_inventory')).toBe('Set up your workspace');
    expect(workspaceSetupDescription('inventory', 'Cabin')).toBe('Name the first inventory for Cabin.');
  });
});
