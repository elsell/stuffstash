import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { CustomAssetType, CustomFieldDefinition, CustomFieldType } from '$lib/domain/inventory';
import AddAssetCustomFieldsSection, { type AddAssetCustomFieldsSectionProps } from './AddAssetCustomFieldsSection.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AddAssetCustomFieldsSection', () => {
  it('renders custom asset types as an accessible pressed choice group', () => {
    const selectedIds: string[] = [];
    component = mount(AddAssetCustomFieldsSection, {
      target: document.body,
      props: sectionProps({
        activeCustomAssetTypes: [customAssetType('tool', 'Tool'), customAssetType('supply', 'Supply')],
        customAssetTypeId: 'tool',
        onCustomAssetTypeSelect: (id) => {
          selectedIds.push(id);
        }
      })
    });

    expect(group('Custom asset type')?.textContent).toContain('Base asset');
    expect(button('Tool').getAttribute('aria-pressed')).toBe('true');
    expect(button('Supply').getAttribute('aria-pressed')).toBe('false');

    button('Base asset').click();

    expect(selectedIds).toEqual(['']);
  });

  it('renders boolean and enum fields as choice controls', () => {
    const changes: Array<[string, string]> = [];
    component = mount(AddAssetCustomFieldsSection, {
      target: document.body,
      props: sectionProps({
        applicableFields: [
          customFieldDefinition('fragile', 'Fragile', 'boolean'),
          customFieldDefinition('condition', 'Condition', 'enum', ['new', 'open'])
        ],
        customFieldValues: { fragile: 'true', condition: 'open' },
        onCustomFieldValueChange: (key, value) => {
          changes.push([key, value]);
        }
      })
    });

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

  it('uses typed inputs for text-like custom fields', () => {
    const changes: Array<[string, string]> = [];
    component = mount(AddAssetCustomFieldsSection, {
      target: document.body,
      props: sectionProps({
        applicableFields: [
          customFieldDefinition('sku', 'SKU', 'text'),
          customFieldDefinition('quantity', 'Quantity', 'number'),
          customFieldDefinition('expires', 'Expires', 'date'),
          customFieldDefinition('manual', 'Manual', 'url')
        ],
        customFieldValues: { quantity: '4' },
        onCustomFieldValueChange: (key, value) => {
          changes.push([key, value]);
        }
      })
    });

    expect(input('custom-field-sku').type).toBe('text');
    expect(input('custom-field-quantity').type).toBe('number');
    expect(input('custom-field-expires').type).toBe('date');
    expect(input('custom-field-manual').type).toBe('url');

    const quantity = input('custom-field-quantity');
    quantity.value = '6';
    quantity.dispatchEvent(new Event('input', { bubbles: true }));

    expect(changes).toEqual([['quantity', '6']]);
  });
});

function sectionProps(overrides: Partial<AddAssetCustomFieldsSectionProps> = {}): AddAssetCustomFieldsSectionProps {
  return {
    activeCustomAssetTypes: [],
    applicableFields: [],
    customAssetTypeId: '',
    customFieldValues: {},
    onCustomAssetTypeSelect: () => {},
    onCustomFieldValueChange: () => {},
    ...overrides
  };
}

function customAssetType(id: string, displayName: string): CustomAssetType {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    scope: 'inventory',
    key: id,
    displayName,
    description: '',
    lifecycleState: 'active'
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
