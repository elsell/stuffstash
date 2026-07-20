import { describe, expect, it } from 'vitest';
import { customizationEditorTarget } from './CustomizationRoutesPresentation';

describe('customization routes', () => {
  it('keeps a household manager in inventory read-only detail before an explicit manage action', () => {
    expect(customizationEditorTarget('field', 'inventory', 'field-1', 'active', true, true)).toEqual({
      pathname: '/settings/inventory/fields/[resourceId]',
      params: { resourceId: 'field-1', lifecycle: 'active', inherited: 'true' }
    });
  });

  it('keeps an inherited row read-only in inventory settings without household management access', () => {
    expect(customizationEditorTarget('asset-type', 'inventory', 'type-1', 'archived', true, false)).toEqual({
      pathname: '/settings/inventory/asset-types/[resourceId]',
      params: { resourceId: 'type-1', lifecycle: 'archived', inherited: 'true' }
    });
  });

  it('routes local tags to the inventory editor', () => {
    expect(customizationEditorTarget('tag', 'inventory', 'tag-1', 'active', false, false).pathname)
      .toBe('/settings/inventory/tags/[resourceId]');
  });
});
