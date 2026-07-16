<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import Activity from '@lucide/svelte/icons/activity';
  import Boxes from '@lucide/svelte/icons/boxes';
  import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
  import Shield from '@lucide/svelte/icons/shield';
  import Users from '@lucide/svelte/icons/users';
  import type { Component } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import {
    type AccessInvitationRouteAction,
    type CustomizationRouteAction,
    type SettingsSection
  } from '$lib/application/workspaceRoute';
  import {
    type SettingsSectionIcon,
    settingsAdministrationPresentation,
    settingsOverviewPresentation,
    settingsSectionOptions,
    settingsShellPresentation
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

  let canConfigureTenant = $derived(hasAccessPermission(tenant?.access, 'configure'));
  let canEditAssets = $derived(canEditAsset(inventory));

  const sectionIconComponents: Record<SettingsSectionIcon, Component> = {
    activity: Activity,
    boxes: Boxes,
    sliders: SlidersHorizontal,
    users: Users
  };

  let sectionOptions = $derived(settingsSectionOptions({
      tenantId: tenant?.id ?? inventory?.tenantId ?? null,
      inventoryId: inventory?.id ?? null,
      section,
      invitationStatus,
      auditScope
    }));
  let activeSection = $derived(
    sectionOptions.find((option) => option.current) ??
      (section === 'administration'
        ? { label: 'Administration', description: 'No web administration actions are available' }
        : sectionOptions[0])
  );
  let shellPresentation = $derived(settingsShellPresentation({ tenant, inventory, activeSection }));
  let overviewPresentation = $derived(settingsOverviewPresentation({
    tenantName: tenant?.name ?? null,
    inventoryCount,
    accessRelationship: inventory?.access.relationship ?? '',
    canEditAssets,
    contextLabel: shellPresentation.overviewContextLabel
  }));
  let administrationPresentation = $derived(settingsAdministrationPresentation({ canConfigureTenant }));

  function selectSection(event: MouseEvent, nextSection: SettingsSection): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onSectionChange(nextSection);
  }
</script>

<section class="workspace-main" aria-labelledby="settings-title">
  <div class="section-heading settings-heading">
    <div>
      <h1 id="settings-title">{shellPresentation.title}</h1>
      <p>{shellPresentation.contextLabel}</p>
    </div>
  </div>

  {#if !inventory}
    <div class="empty-state spacious">
      <h2>{shellPresentation.emptyState?.title}</h2>
      <p>{shellPresentation.emptyState?.message}</p>
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
        <p class="visually-hidden" aria-live="polite">{shellPresentation.liveAnnouncement}</p>
      {#if section === 'overview'}
      <section class="settings-panel" aria-labelledby="settings-overview">
        <div class="settings-panel-heading">
          <Boxes aria-hidden="true" />
          <div>
            <h2 id="settings-overview">{overviewPresentation.title}</h2>
            <p>{overviewPresentation.contextLabel}</p>
          </div>
        </div>
        <dl class="detail-list">
          {#each overviewPresentation.rows as row}
            <div><dt>{row.label}</dt><dd>{row.value}</dd></div>
          {/each}
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
            <h2 id="settings-admin">{administrationPresentation.title}</h2>
            <p>{administrationPresentation.description}</p>
          </div>
        </div>
        <p class="muted-note">Return to Overview, Access, Fields, or Activity to manage this inventory.</p>
      </section>
      {/if}
      </div>
    </div>
  {/if}
</section>
