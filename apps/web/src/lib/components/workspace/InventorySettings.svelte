<script lang="ts">
  import Boxes from '@lucide/svelte/icons/boxes';
  import Shield from '@lucide/svelte/icons/shield';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { SettingsSection } from '$lib/application/workspaceRoute';
  import type { CustomAssetType, CustomFieldDefinition, Inventory, Tenant } from '$lib/domain/inventory';
  import { canEditAsset, hasAccessPermission } from '$lib/domain/inventory';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import InventoryAccessManager from './InventoryAccessManager.svelte';
  import InventoryAuditPanel from './InventoryAuditPanel.svelte';
  import InventoryCustomizationManager from './InventoryCustomizationManager.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenant,
    inventory,
    inventoryCount,
    accessRepository,
    auditRepository,
    customizationRepository,
    customAssetTypes,
    customFieldDefinitions,
    section = 'overview',
    onSectionChange,
    onCustomizationChange
  }: {
    tenant: Tenant | null;
    inventory: Inventory | null;
    inventoryCount: number;
    accessRepository: InventoryAccessRepository;
    auditRepository: InventoryAuditRepository;
    customizationRepository: InventoryCustomizationRepository;
    customAssetTypes: CustomAssetType[];
    customFieldDefinitions: CustomFieldDefinition[];
    section?: SettingsSection;
    onSectionChange: (section: SettingsSection) => void;
    onCustomizationChange: (assetTypes: CustomAssetType[], fieldDefinitions: CustomFieldDefinition[]) => void;
  } = $props();

  let canShare = $derived(hasAccessPermission(inventory?.access, 'share'));
  let canConfigureInventory = $derived(hasAccessPermission(inventory?.access, 'configure'));
  let canConfigureTenant = $derived(hasAccessPermission(tenant?.access, 'configure'));
  let canEditAssets = $derived(canEditAsset(inventory));
  const sectionOptions = [
    { value: 'overview', label: 'Overview' },
    { value: 'access', label: 'Access' },
    { value: 'fields', label: 'Fields' },
    { value: 'activity', label: 'Activity' },
    { value: 'administration', label: 'Admin' }
  ];
</script>

<section class="workspace-main" aria-labelledby="settings-title">
  <div class="section-heading">
    <div>
      <h1 id="settings-title">Inventory settings</h1>
      <p>{inventory?.name ?? 'No inventory selected'}</p>
    </div>
    {#if inventory}
      <Badge variant={canConfigureInventory ? 'secondary' : 'outline'}>{inventory.access.relationship}</Badge>
    {/if}
  </div>

  {#if !inventory}
    <div class="empty-state spacious">
      <h2>No inventory selected</h2>
      <p>Select or create an inventory before managing settings.</p>
    </div>
  {:else}
    <div class="settings-shell">
      <SegmentedControl
        label="Settings section"
        value={section}
        options={sectionOptions}
        onSelect={(value) => onSectionChange(value as SettingsSection)}
      />

      <div class="settings-content">
      {#if section === 'overview'}
      <section class="settings-panel" aria-labelledby="settings-overview">
        <div class="settings-panel-heading">
          <Boxes aria-hidden="true" />
          <div>
            <h2 id="settings-overview">Overview</h2>
            <p>{tenant?.name ?? 'No tenant'} / {inventory.name}</p>
          </div>
        </div>
        <dl class="detail-list">
          <div><dt>Tenant</dt><dd>{tenant?.name ?? 'Not available'}</dd></div>
          <div><dt>Inventories</dt><dd>{inventoryCount}</dd></div>
          <div><dt>Access</dt><dd>{inventory.access.relationship}</dd></div>
          <div><dt>Asset edits</dt><dd>{canEditAssets ? 'Allowed' : 'View only'}</dd></div>
        </dl>
      </section>

      {:else if section === 'access'}
      <InventoryAccessManager {tenant} {inventory} repository={accessRepository} />

      {:else if section === 'activity'}
      <InventoryAuditPanel {tenant} {inventory} repository={auditRepository} />

      {:else if section === 'fields'}
      <InventoryCustomizationManager
        {tenant}
        {inventory}
        repository={customizationRepository}
        initialAssetTypes={customAssetTypes}
        initialFieldDefinitions={customFieldDefinitions}
        onSchemaChange={onCustomizationChange}
      />

      {:else}
      <section class="settings-panel" aria-labelledby="settings-admin">
        <div class="settings-panel-heading">
          <Shield aria-hidden="true" />
          <div>
            <h2 id="settings-admin">Administration</h2>
            <p>
              {canConfigureTenant
                ? 'Tenant-level administration is planned for this workspace.'
                : 'Tenant administration is not available for this account.'}
            </p>
          </div>
        </div>
        <Button.Root variant="outline" disabled={true}>Tenant administration unavailable</Button.Root>
      </section>
      {/if}
      </div>
    </div>
  {/if}
</section>
