<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import './settings-management.css';
  import { inventoryExportContext } from '$lib/ports/inventoryExport';
  import type { ExportInventory } from '$lib/application/exportInventory';
  import InventoryExportAction from './InventoryExportAction.svelte';
  const exportCommand = getContext<ExportInventory | undefined>(inventoryExportContext);
  import { notificationWorkspaceContext, type NotificationWorkspace } from '$lib/ports/notificationWorkspace';
  import NotificationSettings from './NotificationSettings.svelte';
  import { getContext } from 'svelte';
  import { conversationWorkspaceContext, type ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
  import ConversationWorkspace from './conversations/ConversationWorkspace.svelte';
  const notifications = getContext<NotificationWorkspace | undefined>(notificationWorkspaceContext);
  const conversations = getContext<ConversationWorkspaceRepositories | undefined>(conversationWorkspaceContext);
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import { settingsOverviewDestinations, tenantSettingsDestinations, inventorySettingsDestinations, settingsResourceHref } from '$lib/application/settingsManagementNavigation';
  import type { WorkspaceRouteState } from '$lib/application/workspaceRoute';
  import type { CustomAssetType, CustomFieldDefinition, Inventory, ManagedAssetTag, Principal, Tenant } from '$lib/domain/inventory';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import type { InventoryTagRepository } from '$lib/ports/inventoryTagRepository';
  import * as Button from '$lib/components/ui/button/index.js';
  import InventoryAccessManager from '../InventoryAccessManager.svelte';
  import InventoryAuditPanel from '../InventoryAuditPanel.svelte';
  import SettingsDestinationList from './SettingsDestinationList.svelte';
  import TagSettingsManager from './TagSettingsManager.svelte';
  import AssetTypeSettingsManager from './AssetTypeSettingsManager.svelte';
  import FieldSettingsManager from './FieldSettingsManager.svelte';
  import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';

  type Repository = InventoryAccessRepository & InventoryAuditRepository & InventoryCustomizationRepository & InventoryTagRepository;
  type SettingsRoute = Pick<WorkspaceRouteState, 'settingsLevel' | 'settingsCollection' | 'settingsLifecycle' | 'settingsResourceId' | 'settingsResourceAction' | 'invitationStatus' | 'accessInvitationAction' | 'accessInvitationId' | 'auditScope'>;
  let { principal, tenant, inventory, route, repository, observer, currentAssetTypes, currentFields, onNavigate, onSchemaChange, onTagsChange, onPermissionDenied }:
    { principal: Principal; tenant: Tenant | null; inventory: Inventory | null; route: SettingsRoute; repository: Repository; observer: WorkspaceObserver; currentAssetTypes: CustomAssetType[]; currentFields: CustomFieldDefinition[]; onNavigate: (href: string) => void; onSchemaChange: (assetTypes: CustomAssetType[], fields: CustomFieldDefinition[]) => void; onTagsChange: (tags: ManagedAssetTag[]) => void; onPermissionDenied: () => Promise<void> } = $props();

  let latestTypes: CustomAssetType[] = $state([]);
  let latestFields: CustomFieldDefinition[] = $state([]);
  let observedRoute = '';
  let levelTitle = $derived(route.settingsLevel === 'tenant' ? tenant?.name : route.settingsLevel === 'inventory' ? inventory?.name : 'Settings');
  let levelLabel = $derived(route.settingsLevel === 'tenant' ? 'Tenant settings' : route.settingsLevel === 'inventory' ? 'Inventory settings' : '');
  let levelHref = $derived(route.settingsLevel === 'tenant' && tenant ? settingsResourceHref({ level: 'tenant', tenantId: tenant.id }) : tenant && inventory ? settingsResourceHref({ level: 'inventory', tenantId: tenant.id, inventoryId: inventory.id }) : '/settings');
  function navigate(event: MouseEvent, href: string): void { event.preventDefault(); onNavigate(href); }
  function updateTypes(types: CustomAssetType[]): void { latestTypes = types; onSchemaChange(latestTypes, latestFields.length ? latestFields : currentFields); }
  function updateFields(fields: CustomFieldDefinition[]): void { latestFields = fields; onSchemaChange(latestTypes.length ? latestTypes : currentAssetTypes, latestFields); }
  function accessHref(status = route.invitationStatus, action = route.accessInvitationAction, invitationId = route.accessInvitationId): string {
    return settingsResourceHref({ level: 'inventory', tenantId: tenant!.id, inventoryId: inventory!.id, collection: 'access', invitationStatus: status, accessInvitationAction: action, accessInvitationId: invitationId ?? undefined });
  }
  function activityHref(scope = route.auditScope): string {
    return settingsResourceHref({ level: 'inventory', tenantId: tenant!.id, inventoryId: inventory!.id, collection: 'activity', auditScope: scope });
  }
  $effect(() => { const key = `${route.settingsLevel}:${route.settingsCollection ?? 'overview'}`; if (key === observedRoute) return; observedRoute = key; observer.record('workspace.settings_opened', { level: route.settingsLevel, collection: route.settingsCollection ?? 'overview' }); });
</script>

{#if route.settingsLevel === 'overview'}
  <section class="workspace-main settings-management" aria-labelledby="settings-management-title">
    <header class="settings-management-heading"><h1 id="settings-management-title">{t('web.SettingsWorkspace.settings')}</h1><p>{t('web.SettingsWorkspace.chooseWhatYouWantToConfigure')}</p></header>
    <SettingsDestinationList label={t('web.SettingsWorkspace.settingsLevels')} destinations={settingsOverviewDestinations({ tenant, inventory })} {onNavigate} />
  </section>
{:else if route.settingsLevel === 'account'}
  <section class="workspace-main settings-management" aria-labelledby="account-settings-title">
    <Button.Root href="/settings" variant="ghost" class="settings-back" onclick={(event) => navigate(event, '/settings')}><ArrowLeft /> {t('web.SettingsWorkspace.settings')}</Button.Root>
    <header class="settings-management-heading"><p class="settings-eyebrow">{t('web.SettingsWorkspace.accountAndApp')}</p><h1 id="account-settings-title">{t('web.SettingsWorkspace.account')}</h1><p>{principal.email ?? t('web.SettingsWorkspace.signedInAccount')}</p></header>
    <dl class="settings-readonly-details"><div><dt>{t('web.SettingsWorkspace.profileEditing')}</dt><dd>{t('web.SettingsWorkspace.notAvailable')}</dd></div><div><dt>{t('web.SettingsWorkspace.app')}</dt><dd>{t('web.SettingsWorkspace.stuffStashWeb')}</dd></div></dl>
  </section>
{:else if !tenant || (route.settingsLevel === 'inventory' && !inventory)}
  <section class="workspace-main settings-management"><div class="settings-collection-state" role="alert"><h1>{t('web.SettingsWorkspace.settingsUnavailable')}</h1><p>{t('web.SettingsWorkspace.theSelectedSettingsContextIsNotAvailableToThis')}</p><Button.Root href="/settings" onclick={(event) => navigate(event, '/settings')}>{t('web.SettingsWorkspace.backToSettings')}</Button.Root></div></section>
{:else if !route.settingsCollection}
  <section class="workspace-main settings-management" aria-labelledby="settings-level-title">
    <Button.Root href="/settings" variant="ghost" class="settings-back" onclick={(event) => navigate(event, '/settings')}><ArrowLeft /> {t('web.SettingsWorkspace.settings')}</Button.Root>
    <header class="settings-management-heading"><p class="settings-eyebrow">{levelLabel}</p><h1 id="settings-level-title">{levelTitle}</h1>{#if inventory}<p>{t('web.SettingsWorkspace.belongsToFull', { name: inventory.name, name2: tenant.name })}</p>{:else}<p>{t('web.SettingsWorkspace.settingsSharedWithThisTenantSInventories')}</p>{/if}</header>
    <SettingsDestinationList label={t('web.SettingsWorkspace.settings2', { levelTitle: String(levelTitle) })} destinations={route.settingsLevel === 'tenant' ? tenantSettingsDestinations(tenant) : inventorySettingsDestinations(inventory!)} {onNavigate} />
    {#if route.settingsLevel === 'inventory' && inventory && exportCommand}
      {#key JSON.stringify([principal.id, tenant.id, inventory.id])}
        <InventoryExportAction command={exportCommand} scope={{ tenantId: tenant.id, inventoryId: inventory.id }} />
      {/key}
    {/if}
  </section>
{:else}
  <div class="workspace-main settings-management settings-management-resource">
    <Button.Root href={levelHref} variant="ghost" class="settings-back" onclick={(event) => navigate(event, levelHref)}><ArrowLeft /> {levelTitle}</Button.Root>
    {#if route.settingsCollection === 'notifications' && inventory && notifications}
      {#key JSON.stringify([notifications.apiIdentity, principal.id, tenant.id, inventory.id])}
        <NotificationSettings tenantId={tenant.id} inventoryId={inventory.id} initialTimezone={Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'} repository={notifications.repository} onChanged={notifications.onPreferencesChanged} {observer} typeRepository={repository} />
      {/key}
    {:else if route.settingsCollection === 'conversations' && route.settingsLevel === 'tenant' && conversations}
      {#key JSON.stringify([conversations.apiIdentity, principal.id, tenant.id])}
        <ConversationWorkspace scope={{ apiIdentity: conversations.apiIdentity, principalId: principal.id, tenantId: tenant.id }} repositories={conversations} />
      {/key}
    {:else if route.settingsCollection === 'tags' && inventory}
      <TagSettingsManager {inventory} {repository} {observer} resourceId={route.settingsResourceId} action={route.settingsResourceAction} {onNavigate} {onTagsChange} {onPermissionDenied} />
    {:else if route.settingsCollection === 'asset-types'}
      <AssetTypeSettingsManager level={route.settingsLevel as 'tenant' | 'inventory'} {tenant} {inventory} {repository} {observer} canonicalItems={currentAssetTypes} lifecycle={route.settingsLifecycle} resourceId={route.settingsResourceId} action={route.settingsResourceAction} {onNavigate} onSchemaChange={updateTypes} {onPermissionDenied} />
    {:else if route.settingsCollection === 'fields'}
      <FieldSettingsManager level={route.settingsLevel as 'tenant' | 'inventory'} {tenant} {inventory} {repository} {observer} canonicalItems={currentFields} lifecycle={route.settingsLifecycle} resourceId={route.settingsResourceId} action={route.settingsResourceAction} {onNavigate} onSchemaChange={updateFields} {onPermissionDenied} />
    {:else if route.settingsCollection === 'access' && inventory}
      <InventoryAccessManager {tenant} {inventory} {repository} invitationStatus={route.invitationStatus} accessInvitationAction={route.accessInvitationAction} accessInvitationId={route.accessInvitationId} onInvitationStatusChange={(status) => onNavigate(accessHref(status, null, null))} onInvitationActionOpen={(action, invitationId) => onNavigate(accessHref(route.invitationStatus, action, invitationId))} onInvitationActionClose={() => onNavigate(accessHref(route.invitationStatus, null, null))} />
    {:else if route.settingsCollection === 'activity' && inventory}
      <InventoryAuditPanel {tenant} {inventory} {repository} scope={route.auditScope} onScopeChange={(scope) => onNavigate(activityHref(scope))} />
    {:else}
      <div class="settings-collection-state" role="alert"><h1>{t('web.SettingsWorkspace.sectionUnavailable')}</h1><p>{t('web.SettingsWorkspace.thisSettingsSectionIsNotAvailableInTheSelected')}</p></div>
    {/if}
  </div>
{/if}
