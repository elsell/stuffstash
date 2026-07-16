import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import InventoryCustomizationManager from './InventoryCustomizationManager.svelte';
import type { CustomAssetType, CustomFieldDefinition, Inventory, Tenant } from '$lib/domain/inventory';
import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';

let component: ReturnType<typeof mount> | null = null;

afterEach(async () => {
  document.body.querySelector<HTMLElement>('[role="alertdialog"]')?.dispatchEvent(
    new KeyboardEvent('keydown', { key: 'Escape', bubbles: true })
  );
  await new Promise((resolve) => window.setTimeout(resolve, 20));
  if (component) {
    await unmount(component);
    document.body.innerHTML = '';
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryCustomizationManager', () => {
  it('keeps asset types and field definitions in independently sized grouped surfaces', async () => {
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository: fakeCustomizationRepository(),
        initialAssetTypes: [],
        initialFieldDefinitions: [],
        onSchemaChange: () => {}
      }
    });
    await flush();

    const surfaces = document.body.querySelectorAll('section.customization-surface');
    expect(surfaces).toHaveLength(2);
    expect(surfaces[0]?.querySelector('h3')?.textContent).toBe('Asset types');
    expect(surfaces[1]?.querySelector('h3')?.textContent).toBe('Field definitions');
    expect(surfaces[0]?.querySelector('[aria-label="0 custom asset types"]')?.textContent).toBe('0');
    expect(surfaces[1]?.querySelector('[aria-label="0 custom fields"]')?.textContent).toBe('0');
    expect(surfaces[0]?.textContent).toContain('No custom asset types yet.');
    expect(surfaces[1]?.textContent).toContain('No custom fields yet.');
  });

  it('counts only active definitions and removes empty copy when rows are visible', async () => {
    const medicine = customAssetType();
    const expiration = customFieldDefinition();
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository: fakeCustomizationRepository(),
        initialAssetTypes: [medicine, { ...medicine, id: 'type-archived', key: 'archived', lifecycleState: 'archived' }],
        initialFieldDefinitions: [
          expiration,
          { ...expiration, id: 'field-archived', key: 'archived-field', lifecycleState: 'archived' }
        ],
        onSchemaChange: () => {}
      }
    });
    await flush();

    const surfaces = document.body.querySelectorAll('section.customization-surface');
    expect(surfaces[0]?.querySelector('[aria-label="1 custom asset types"]')?.textContent).toBe('1');
    expect(surfaces[1]?.querySelector('[aria-label="1 custom fields"]')?.textContent).toBe('1');
    expect(surfaces[0]?.textContent).toContain('Medicine');
    expect(surfaces[1]?.textContent).toContain('Expiration date');
    expect(surfaces[0]?.textContent).not.toContain('No custom asset types yet.');
    expect(surfaces[1]?.textContent).not.toContain('No custom fields yet.');
  });

  it('renders missing inventory customization status without an alert role', async () => {
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: null,
        repository: fakeCustomizationRepository(),
        initialAssetTypes: [],
        initialFieldDefinitions: [],
        onSchemaChange: () => {}
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Select an inventory before managing fields.');
    expect(document.body.querySelector('[role="alert"]')).toBeNull();
    expect(exactButtonOrNull('Create type')).toBeNull();
  });

  it('renders denied customization status as an alert', async () => {
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: { ...tenant(), access: { relationship: 'viewer', permissions: ['view'] } },
        inventory: { ...inventory(), access: { relationship: 'viewer', permissions: ['view'] } },
        repository: fakeCustomizationRepository(),
        initialAssetTypes: [],
        initialFieldDefinitions: [],
        onSchemaChange: () => {}
      }
    });
    await flush();

    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain(
      'Custom fields require tenant or inventory configuration access.'
    );
    expect(exactButtonOrNull('Create type')).toBeNull();
  });

  it('renders customization operation failures as alerts while preserving the form', async () => {
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository: fakeCustomizationRepository(),
        initialAssetTypes: [],
        initialFieldDefinitions: [],
        onSchemaChange: () => {}
      }
    });
    await flush();

    input('#custom-type-key', 'medicine');
    input('#custom-type-name', 'Medicine');
    await flush();
    clickExactButton('Create type');
    await flush();

    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain('Custom asset type could not be created. Try again.');
    expect(document.body.textContent).not.toContain('Unexpected repository call.');
    expect(exactButton('Create type')).toBeTruthy();
  });

  it('creates custom asset types and targeted custom fields through the customization port', async () => {
    const calls: string[] = [];
    const medicine: CustomAssetType = {
      id: 'type-medicine',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      scope: 'inventory',
      key: 'medicine',
      displayName: 'Medicine',
      description: 'Medication',
      lifecycleState: 'active'
    };
    const schemaChanges: string[] = [];
    const repository: InventoryCustomizationRepository = {
      listInventoryCustomAssetTypes: async () => ({ items: [], pagination: page() }),
      createCustomAssetType: async (_tenantId, _inventoryId, draft) => {
        calls.push(`type:${draft.scope}:${draft.key}:${draft.displayName}`);
        return { ...medicine, key: draft.key, displayName: draft.displayName, description: draft.description };
      },
      archiveCustomAssetType: async () => medicine,
      listInventoryCustomFieldDefinitions: async () => ({ items: [], pagination: page() }),
      createCustomFieldDefinition: async (_tenantId, _inventoryId, draft) => {
        calls.push(`field:${draft.scope}:${draft.key}:${draft.type}:${draft.applicability}:${draft.customAssetTypeIds.join(',')}`);
        return {
          id: 'field-expiration',
          tenantId: 'tenant-one',
          inventoryId: 'inventory-one',
          scope: draft.scope,
          key: draft.key,
          displayName: draft.displayName,
          type: draft.type,
          enumOptions: draft.enumOptions,
          applicability: draft.applicability,
          customAssetTypeIds: draft.customAssetTypeIds,
          lifecycleState: 'active'
        };
      },
      archiveCustomFieldDefinition: async () => ({
        id: 'field-expiration',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        scope: 'inventory',
        key: 'expiration-date',
        displayName: 'Expiration date',
        type: 'date',
        enumOptions: [],
        applicability: 'custom_asset_types',
        customAssetTypeIds: ['type-medicine'],
        lifecycleState: 'archived'
      })
    };

    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository,
        initialAssetTypes: [medicine],
        initialFieldDefinitions: [],
        onSchemaChange: (assetTypes, fieldDefinitions) => {
          schemaChanges.push(`${assetTypes.length}:${fieldDefinitions.length}`);
        }
      }
    });
    await flush();

    input('#custom-type-key', 'medicine');
    input('#custom-type-name', 'Medicine');
    input('#custom-type-description', 'Medication');
    await flush();
    click('Create type');
    await flush();
    click('Custom types');
    await flush();
    input('#custom-field-key', 'expiration-date');
    input('#custom-field-name', 'Expiration date');
    click('Date');
    click('Medicine');
    await flush();
    expect(group('Field custom type targets')?.querySelector('button[aria-pressed="true"]')?.textContent).toContain('Medicine');
    expect(document.body.textContent).toContain('1 custom type selected');
    click('Create field');
    await flush();

    expect(calls).toEqual([
      'type:inventory:medicine:Medicine',
      'field:inventory:expiration-date:date:custom_asset_types:type-medicine'
    ]);
    expect(schemaChanges).toEqual(['2:0', '2:1']);
  });

  it('shows a calm empty state when no custom types are eligible field targets', async () => {
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository: fakeCustomizationRepository(),
        initialAssetTypes: [],
        initialFieldDefinitions: [],
        onSchemaChange: () => {}
      }
    });
    await flush();

    click('Custom types');
    await flush();

    expect(document.body.textContent).toContain('No custom types selected');
    expect(document.body.textContent).toContain('No eligible custom asset types for this scope.');
    expect(group('Field custom type targets')).toBeNull();
  });

  it('selects the tenant scope when only tenant configuration access is available', async () => {
    const calls: string[] = [];
    const repository: InventoryCustomizationRepository = {
      ...fakeCustomizationRepository(),
      createCustomAssetType: async (_tenantId, _inventoryId, draft) => {
        calls.push(`type:${draft.scope}:${draft.key}`);
        return {
          id: 'type-appliance',
          tenantId: 'tenant-one',
          inventoryId: null,
          scope: draft.scope,
          key: draft.key,
          displayName: draft.displayName,
          description: draft.description,
          lifecycleState: 'active'
        };
      },
      createCustomFieldDefinition: async (_tenantId, _inventoryId, draft) => {
        calls.push(`field:${draft.scope}:${draft.key}`);
        return {
          id: 'field-warranty',
          tenantId: 'tenant-one',
          inventoryId: null,
          scope: draft.scope,
          key: draft.key,
          displayName: draft.displayName,
          type: draft.type,
          enumOptions: draft.enumOptions,
          applicability: draft.applicability,
          customAssetTypeIds: draft.customAssetTypeIds,
          lifecycleState: 'active'
        };
      }
    };

    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: { ...inventory(), access: { relationship: 'viewer', permissions: ['view'] } },
        repository,
        initialAssetTypes: [],
        initialFieldDefinitions: [],
        onSchemaChange: () => {}
      }
    });
    await flush();

    expect(selectedOption('Custom type scope')?.textContent).toContain('Tenant');
    expect(selectedOption('Custom field scope')?.textContent).toContain('Tenant');
    expect(buttonInGroup('Custom type scope', 'Inventory')?.hasAttribute('disabled')).toBe(true);
    expect(buttonInGroup('Custom type scope', 'Tenant')?.hasAttribute('disabled')).toBe(false);

    input('#custom-type-key', 'appliance');
    input('#custom-type-name', 'Appliance');
    input('#custom-field-key', 'warranty');
    input('#custom-field-name', 'Warranty');
    await flush();

    expect(exactButton('Create type').hasAttribute('disabled')).toBe(false);
    expect(exactButton('Create field').hasAttribute('disabled')).toBe(false);

    clickExactButton('Create type');
    clickExactButton('Create field');
    await flush();

    expect(calls).toEqual(['type:tenant:appliance', 'field:tenant:warranty']);
  });

  it('uses route-backed archive confirmations for custom schema actions', async () => {
    const medicine = customAssetType();
    const expiration = customFieldDefinition();
    const calls: string[] = [];
    const openedActions: string[] = [];
    let closed = 0;
    const repository: InventoryCustomizationRepository = {
      ...fakeCustomizationRepository(),
      archiveCustomAssetType: async (_tenantId, _inventoryId, id, scope) => {
        calls.push(`archive-type:${id}:${scope}`);
        return { ...medicine, lifecycleState: 'archived' };
      }
    };

    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository,
        initialAssetTypes: [medicine],
        initialFieldDefinitions: [expiration],
        onArchiveActionOpen: (action, id) => {
          openedActions.push(`${action}:${id}`);
        },
        onArchiveActionClose: () => {
          closed += 1;
        },
        onSchemaChange: () => {}
      }
    });
    await flush();

    expect(controlWithLabel('Archive Medicine').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/fields/asset-types/type-medicine/archive'
    );
    expect(controlWithLabel('Archive Expiration date').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/fields/field-definitions/field-expiration/archive'
    );

    controlWithLabel('Archive Medicine').click();
    await flush();

    expect(openedActions).toEqual(['archive_asset_type:type-medicine']);
    expect(calls).toEqual([]);

    await unmount(component);
    component = null;
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository,
        initialAssetTypes: [medicine],
        initialFieldDefinitions: [expiration],
        archiveAction: 'archive_asset_type',
        archiveAssetTypeId: medicine.id,
        onArchiveActionClose: () => {
          closed += 1;
        },
        onSchemaChange: () => {}
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Archive asset type');
    expect(document.activeElement?.textContent).toBe('Cancel');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/fields');

    clickExactButton('Archive');
    await flush();
    await waitForDialogClose(() => closed === 1);

    expect(calls).toEqual(['archive-type:type-medicine:inventory']);
    expect(closed).toBe(1);
  });

  it('renders and applies route-backed custom field archive confirmations', async () => {
    const medicine = customAssetType();
    const expiration = customFieldDefinition();
    const calls: string[] = [];
    let closed = 0;
    const repository: InventoryCustomizationRepository = {
      ...fakeCustomizationRepository(),
      archiveCustomFieldDefinition: async (_tenantId, _inventoryId, id, scope) => {
        calls.push(`archive-field:${id}:${scope}`);
        return { ...expiration, lifecycleState: 'archived' };
      }
    };

    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository,
        initialAssetTypes: [medicine],
        initialFieldDefinitions: [expiration],
        archiveAction: 'archive_field_definition',
        archiveFieldDefinitionId: expiration.id,
        onArchiveActionClose: () => {
          closed += 1;
        },
        onSchemaChange: () => {}
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Archive field definition');
    expect(document.activeElement?.textContent).toBe('Cancel');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/fields');

    link('Cancel').click();
    await waitForDialogClose(() => closed === 1);

    expect(closed).toBe(1);
    expect(calls).toEqual([]);

    await unmount(component);
    component = null;
    document.body.innerHTML = '';
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository,
        initialAssetTypes: [medicine],
        initialFieldDefinitions: [expiration],
        archiveAction: 'archive_field_definition',
        archiveFieldDefinitionId: expiration.id,
        onArchiveActionClose: () => {
          closed += 1;
        },
        onSchemaChange: () => {}
      }
    });
    await flush();

    clickExactButton('Archive');
    await flush();
    await waitForDialogClose(() => closed === 2);

    expect(calls).toEqual(['archive-field:field-expiration:inventory']);
    expect(closed).toBe(2);
  });

  it('renders unavailable archive routes from stale schema ids and closes them', async () => {
    let closed = 0;
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository: fakeCustomizationRepository(),
        initialAssetTypes: [customAssetType()],
        initialFieldDefinitions: [customFieldDefinition()],
        archiveAction: 'archive_asset_type',
        archiveAssetTypeId: 'missing-type',
        onArchiveActionClose: () => {
          closed += 1;
        },
        onSchemaChange: () => {}
      }
    });
    await flush();

    expect(document.body.textContent).toContain('Archive target unavailable');
    link('Back to fields').click();
    await waitForDialogClose(() => closed === 1);

    expect(closed).toBe(1);
  });

  it('moves focus to the Custom fields heading when a deep-linked archive confirmation closes', async () => {
    component = mount(InventoryCustomizationManager, {
      target: document.body,
      props: {
        tenant: tenant(),
        inventory: inventory(),
        repository: fakeCustomizationRepository(),
        initialAssetTypes: [customAssetType()],
        initialFieldDefinitions: [customFieldDefinition()],
        archiveAction: 'archive_asset_type',
        archiveAssetTypeId: 'type-medicine',
        onSchemaChange: () => {}
      }
    });
    await flush();

    document.body.querySelector<HTMLElement>('[role="alertdialog"]')?.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Escape', bubbles: true })
    );
    await waitForDialogClose(() => document.activeElement?.id === 'settings-customization');

    expect(document.activeElement).toBe(document.getElementById('settings-customization'));
  });
});

function customAssetType(): CustomAssetType {
  return {
    id: 'type-medicine',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    scope: 'inventory',
    key: 'medicine',
    displayName: 'Medicine',
    description: 'Medication',
    lifecycleState: 'active'
  };
}

function customFieldDefinition(): CustomFieldDefinition {
  return {
    id: 'field-expiration',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    scope: 'inventory',
    key: 'expiration-date',
    displayName: 'Expiration date',
    type: 'date',
    enumOptions: [],
    applicability: 'custom_asset_types',
    customAssetTypeIds: ['type-medicine'],
    lifecycleState: 'active'
  };
}

function tenant(): Tenant {
  return { id: 'tenant-one', name: 'Home', access: { relationship: 'owner', permissions: ['view', 'configure'] } };
}

function inventory(): Inventory {
  return {
    id: 'inventory-one',
    tenantId: 'tenant-one',
    name: 'Household',
    access: { relationship: 'owner', permissions: ['view', 'configure'] }
  };
}

function page() {
  return { limit: 50, nextCursor: null, hasMore: false };
}

function input(selector: string, value: string): void {
  const element = document.querySelector<HTMLInputElement | HTMLTextAreaElement>(selector);
  if (!element) throw new Error(`Missing input ${selector}`);
  element.value = value;
  element.dispatchEvent(new Event('input', { bubbles: true }));
}

function click(text: string): void {
  const button = Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent?.includes(text));
  if (!button) throw new Error(`Missing button ${text}`);
  button.click();
}

function clickExactButton(text: string): void {
  exactButton(text).click();
}

function exactButton(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent === text);
  if (!button) throw new Error(`Missing button ${text}`);
  return button;
}

function exactButtonOrNull(text: string): HTMLButtonElement | null {
  return Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent === text) ?? null;
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent === text);
  if (!target) throw new Error(`Missing link ${text}`);
  return target;
}

function controlWithLabel(label: string): HTMLElement {
  const control = document.body.querySelector<HTMLElement>(`button[aria-label="${label}"], a[aria-label="${label}"]`);
  if (!control) throw new Error(`Missing control labelled ${label}`);
  return control;
}

function group(label: string): HTMLElement | null {
  return document.body.querySelector<HTMLElement>(`[role="group"][aria-label="${label}"]`);
}

function selectedOption(groupLabel: string): HTMLButtonElement | null {
  return group(groupLabel)?.querySelector<HTMLButtonElement>('button[aria-pressed="true"]') ?? null;
}

function buttonInGroup(groupLabel: string, text: string): HTMLButtonElement | null {
  return Array.from(group(groupLabel)?.querySelectorAll<HTMLButtonElement>('button') ?? []).find((button) =>
    button.textContent?.includes(text)
  ) ?? null;
}

function fakeCustomizationRepository(): InventoryCustomizationRepository {
  return {
    listInventoryCustomAssetTypes: async () => ({ items: [], pagination: page() }),
    createCustomAssetType: async () => failRepositoryCall(),
    archiveCustomAssetType: async () => failRepositoryCall(),
    listInventoryCustomFieldDefinitions: async () => ({ items: [], pagination: page() }),
    createCustomFieldDefinition: async () => failRepositoryCall(),
    archiveCustomFieldDefinition: async () => failRepositoryCall()
  };
}

function failRepositoryCall(): never {
  throw new Error('Unexpected repository call.');
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}

async function waitForDialogClose(condition: () => boolean): Promise<void> {
  for (let attempt = 0; attempt < 100; attempt += 1) {
    await flush();
    if (condition() && !document.body.querySelector('[role="alertdialog"]')) {
      return;
    }
    await new Promise((resolve) => window.setTimeout(resolve, 5));
  }
  throw new Error('Dialog did not finish closing.');
}
