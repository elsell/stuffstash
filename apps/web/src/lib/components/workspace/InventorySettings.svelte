<script lang="ts">
  import Activity from '@lucide/svelte/icons/activity';
  import Boxes from '@lucide/svelte/icons/boxes';
  import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
  import Shield from '@lucide/svelte/icons/shield';
  import UserRoundCog from '@lucide/svelte/icons/user-round-cog';
  import Users from '@lucide/svelte/icons/users';
  import type { Component } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { workspaceRouteHref, type CustomizationRouteAction, type SettingsSection } from '$lib/application/workspaceRoute';
  import type { AuditScope, CustomAssetType, CustomFieldDefinition, Inventory, InvitationStatusFilter, Tenant } from '$lib/domain/inventory';
  import { canEditAsset, hasAccessPermission } from '$lib/domain/inventory';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import InventoryAccessManager from './InventoryAccessManager.svelte';
  import InventoryAuditPanel from './InventoryAuditPanel.svelte';
  import InventoryCustomizationManager from './InventoryCustomizationManager.svelte';

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
    invitationStatus = 'all',
    auditScope = 'inventory',
    customizationAction = null,
    customAssetTypeId = null,
    customFieldDefinitionId = null,
    onSectionChange,
    onInvitationStatusChange,
    onAuditScopeChange,
    onCustomizationArchiveOpen = () => {},
    onCustomizationArchiveClose = () => {},
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
    invitationStatus?: InvitationStatusFilter;
    auditScope?: AuditScope;
    customizationAction?: CustomizationRouteAction;
    customAssetTypeId?: string | null;
    customFieldDefinitionId?: string | null;
    onSectionChange: (section: SettingsSection) => void;
    onInvitationStatusChange: (status: InvitationStatusFilter) => void;
    onAuditScopeChange: (scope: AuditScope) => void;
    onCustomizationArchiveOpen?: (action: CustomizationRouteAction, id: string) => void;
    onCustomizationArchiveClose?: () => void;
    onCustomizationChange: (assetTypes: CustomAssetType[], fieldDefinitions: CustomFieldDefinition[]) => void;
  } = $props();

  let canShare = $derived(hasAccessPermission(inventory?.access, 'share'));
  let canConfigureInventory = $derived(hasAccessPermission(inventory?.access, 'configure'));
  let canConfigureTenant = $derived(hasAccessPermission(tenant?.access, 'configure'));
  let canEditAssets = $derived(canEditAsset(inventory));

  type SectionOption = {
    value: SettingsSection;
    label: string;
    description: string;
    icon: Component;
  };

  const sectionOptions: SectionOption[] = [
    { value: 'overview', label: 'Overview', description: 'Inventory context and access summary', icon: Boxes },
    { value: 'access', label: 'Access', description: 'Sharing, grants, and invitations', icon: Users },
    { value: 'fields', label: 'Fields', description: 'Custom asset types and fields', icon: SlidersHorizontal },
    { value: 'activity', label: 'Activity', description: 'Audit history for this workspace', icon: Activity },
    { value: 'administration', label: 'Admin', description: 'Tenant and inventory administration', icon: UserRoundCog }
  ];

  let activeSection = $derived(sectionOptions.find((option) => option.value === section) ?? sectionOptions[0]);

  function sectionHref(nextSection: SettingsSection): string {
    return workspaceRouteHref(
      {
        mode: 'settings',
        settingsSection: nextSection,
        invitationStatus: nextSection === 'access' ? invitationStatus : 'all',
        auditScope: nextSection === 'activity' ? auditScope : 'inventory'
      },
      tenant?.id ?? inventory?.tenantId ?? null,
      inventory?.id ?? null
    );
  }

  function invitationStatusHref(status: InvitationStatusFilter): string {
    return workspaceRouteHref(
      { mode: 'settings', settingsSection: 'access', invitationStatus: status },
      tenant?.id ?? inventory?.tenantId ?? null,
      inventory?.id ?? null
    );
  }

  function auditScopeHref(scope: AuditScope): string {
    return workspaceRouteHref(
      { mode: 'settings', settingsSection: 'activity', auditScope: scope },
      tenant?.id ?? inventory?.tenantId ?? null,
      inventory?.id ?? null
    );
  }

  function selectSection(event: MouseEvent, nextSection: SettingsSection): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onSectionChange(nextSection);
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }
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
      <nav class="settings-section-nav" aria-label="Settings sections">
        {#each sectionOptions as option}
          {@const Icon = option.icon}
          <Button.Root
            href={sectionHref(option.value)}
            variant={section === option.value ? 'secondary' : 'ghost'}
            class="settings-section-link"
            aria-current={section === option.value ? 'page' : undefined}
            onclick={(event) => selectSection(event, option.value)}
          >
            <Icon aria-hidden="true" />
            <span>
              <strong>{option.label}</strong>
              <small>{option.description}</small>
            </span>
          </Button.Root>
        {/each}
      </nav>

      <div class="settings-content">
        <div class="settings-section-context" aria-live="polite">
          <span class="settings-section-kicker">Settings</span>
          <h2>{activeSection.label}</h2>
          <p>{activeSection.description}</p>
        </div>

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
      <InventoryAccessManager
        {tenant}
        {inventory}
        repository={accessRepository}
        {invitationStatus}
        invitationStatusHref={invitationStatusHref}
        onInvitationStatusChange={onInvitationStatusChange}
      />

      {:else if section === 'activity'}
      <InventoryAuditPanel
        {tenant}
        {inventory}
        repository={auditRepository}
        scope={auditScope}
        scopeHref={auditScopeHref}
        onScopeChange={onAuditScopeChange}
      />

      {:else if section === 'fields'}
      <InventoryCustomizationManager
        {tenant}
        {inventory}
        repository={customizationRepository}
        initialAssetTypes={customAssetTypes}
        initialFieldDefinitions={customFieldDefinitions}
        archiveAction={customizationAction}
        archiveAssetTypeId={customAssetTypeId}
        archiveFieldDefinitionId={customFieldDefinitionId}
        onArchiveActionOpen={onCustomizationArchiveOpen}
        onArchiveActionClose={onCustomizationArchiveClose}
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
