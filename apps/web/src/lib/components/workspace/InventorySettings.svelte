<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import Activity from '@lucide/svelte/icons/activity';
  import Boxes from '@lucide/svelte/icons/boxes';
  import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
  import Shield from '@lucide/svelte/icons/shield';
  import UserRoundCog from '@lucide/svelte/icons/user-round-cog';
  import Users from '@lucide/svelte/icons/users';
  import type { Component } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import {
    type AccessInvitationRouteAction,
    type CustomizationRouteAction,
    type SettingsSection
  } from '$lib/application/workspaceRoute';
  import {
    type SettingsSectionIcon,
    settingsSectionOptions
  } from '$lib/application/workspaceSettingsNavigation';
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
    accessInvitationAction = null,
    accessInvitationId = null,
    auditScope = 'inventory',
    customizationAction = null,
    customAssetTypeId = null,
    customFieldDefinitionId = null,
    onSectionChange,
    onInvitationStatusChange,
    onAccessInvitationActionOpen = () => {},
    onAccessInvitationActionClose = () => {},
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
    accessInvitationAction?: AccessInvitationRouteAction;
    accessInvitationId?: string | null;
    auditScope?: AuditScope;
    customizationAction?: CustomizationRouteAction;
    customAssetTypeId?: string | null;
    customFieldDefinitionId?: string | null;
    onSectionChange: (section: SettingsSection) => void;
    onInvitationStatusChange: (status: InvitationStatusFilter) => void;
    onAccessInvitationActionOpen?: (action: AccessInvitationRouteAction, invitationId: string) => void;
    onAccessInvitationActionClose?: () => void;
    onAuditScopeChange: (scope: AuditScope) => void;
    onCustomizationArchiveOpen?: (action: CustomizationRouteAction, id: string) => void;
    onCustomizationArchiveClose?: () => void;
    onCustomizationChange: (assetTypes: CustomAssetType[], fieldDefinitions: CustomFieldDefinition[]) => void;
  } = $props();

  let canShare = $derived(hasAccessPermission(inventory?.access, 'share'));
  let canConfigureInventory = $derived(hasAccessPermission(inventory?.access, 'configure'));
  let canConfigureTenant = $derived(hasAccessPermission(tenant?.access, 'configure'));
  let canEditAssets = $derived(canEditAsset(inventory));

  const sectionIconComponents: Record<SettingsSectionIcon, Component> = {
    activity: Activity,
    boxes: Boxes,
    sliders: SlidersHorizontal,
    'user-cog': UserRoundCog,
    users: Users
  };

  let sectionOptions = $derived(settingsSectionOptions({
      tenantId: tenant?.id ?? inventory?.tenantId ?? null,
      inventoryId: inventory?.id ?? null,
      section,
      invitationStatus,
      auditScope
    }));
  let activeSection = $derived(sectionOptions.find((option) => option.current) ?? sectionOptions[0]);

  function selectSection(event: MouseEvent, nextSection: SettingsSection): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onSectionChange(nextSection);
  }
</script>

<section class="workspace-main" aria-labelledby="settings-title">
  <div class="section-heading">
    <div>
      <h1 id="settings-title">{inventory?.name ?? 'Settings'}</h1>
      <p>{inventory ? `${tenant?.name ?? 'No tenant'} / ${activeSection.label}` : 'No inventory selected'}</p>
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
          {@const Icon = sectionIconComponents[option.icon]}
          <Button.Root
            href={option.href}
            variant={option.current ? 'secondary' : 'ghost'}
            class="settings-section-link"
            aria-current={option.current ? 'page' : undefined}
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
        <p class="visually-hidden" aria-live="polite">{activeSection.label}: {activeSection.description}</p>
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
        {accessInvitationAction}
        {accessInvitationId}
        onInvitationStatusChange={onInvitationStatusChange}
        onInvitationActionOpen={onAccessInvitationActionOpen}
        onInvitationActionClose={onAccessInvitationActionClose}
      />

      {:else if section === 'activity'}
      <InventoryAuditPanel
        {tenant}
        {inventory}
        repository={auditRepository}
        scope={auditScope}
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
