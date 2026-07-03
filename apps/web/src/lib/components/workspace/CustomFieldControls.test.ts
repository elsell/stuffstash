import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { CustomFieldDefinition, CustomFieldType } from '$lib/domain/inventory';
import CustomFieldControls, { type CustomFieldControlsProps } from './CustomFieldControls.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('CustomFieldControls', () => {
  it('renders boolean and enum fields as named choice groups', () => {
    const changes: Array<[string, string]> = [];
    component = mount(CustomFieldControls, {
      target: document.body,
      props: controlsProps({
        fields: [
          customFieldDefinition('fragile', 'Fragile', 'boolean'),
          customFieldDefinition('condition', 'Condition', 'enum', ['new', 'open'])
        ],
        values: { fragile: 'true', condition: 'open' },
        onValueChange: (key, value) => {
          changes.push([key, value]);
        }
      })
    });

    expect(document.body.querySelector('[aria-label="Custom fields"]')).not.toBeNull();
    expect(group('Fragile')?.textContent).toContain('Yes');
    expect(button('Yes').getAttribute('aria-pressed')).toBe('true');
    expect(button('open').getAttribute('aria-pressed')).toBe('true');

    button('No').click();
    button('new').click();

    expect(changes).toEqual([
      ['fragile', 'false'],
      ['condition', 'new']
    ]);
  });

  it('uses the configured id prefix and typed inputs for text-like fields', () => {
    const changes: Array<[string, string]> = [];
    component = mount(CustomFieldControls, {
      target: document.body,
      props: controlsProps({
        idPrefix: 'edit-custom-field',
        fields: [
          customFieldDefinition('sku', 'SKU', 'text'),
          customFieldDefinition('quantity', 'Quantity', 'number'),
          customFieldDefinition('expires', 'Expires', 'date'),
          customFieldDefinition('manual', 'Manual', 'url')
        ],
        values: { quantity: '4' },
        onValueChange: (key, value) => {
          changes.push([key, value]);
        }
      })
    });

    expect(input('edit-custom-field-sku').type).toBe('text');
    expect(input('edit-custom-field-quantity').type).toBe('number');
    expect(input('edit-custom-field-expires').type).toBe('date');
    expect(input('edit-custom-field-manual').type).toBe('url');

    const quantity = input('edit-custom-field-quantity');
    quantity.value = '6';
    quantity.dispatchEvent(new Event('input', { bubbles: true }));

    expect(changes).toEqual([['quantity', '6']]);
  });
});

function controlsProps(overrides: Partial<CustomFieldControlsProps> = {}): CustomFieldControlsProps {
  return {
    fields: [],
    values: {},
    idPrefix: 'custom-field',
    label: 'Custom fields',
    onValueChange: () => {},
    ...overrides
  };
}

function customFieldDefinition(key: string, displayName: string, type: CustomFieldType, enumOptions: string[] = []): CustomFieldDefinition {
  return {
    id: `field-${key}`,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    scope: 'inventory',
    key,
    displayName,
    type,
    enumOptions,
    applicability: 'all_assets',
    customAssetTypeIds: [],
    lifecycleState: 'active'
  };
}

function group(name: string): HTMLElement | null {
  return document.body.querySelector<HTMLElement>(`[role="group"][aria-label="${name}"]`);
}

function button(name: string): HTMLButtonElement {
  const target = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) => candidate.textContent?.trim() === name);
  if (!target) {
    throw new Error(`Missing button ${name}`);
  }
  return target;
}

function input(id: string): HTMLInputElement {
  const target = document.body.querySelector<HTMLInputElement>(`#${id}`);
  if (!target) {
    throw new Error(`Missing input ${id}`);
  }
  return target;
}
