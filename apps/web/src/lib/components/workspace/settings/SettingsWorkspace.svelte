<script lang="ts">
  import './settings-management.css';
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
    <header class="settings-management-heading"><h1 id="settings-management-title">Settings</h1><p>Choose what you want to configure.</p></header>
    <SettingsDestinationList label="Settings levels" destinations={settingsOverviewDestinations({ tenant, inventory })} {onNavigate} />
  </section>
{:else if route.settingsLevel === 'account'}
  <section class="workspace-main settings-management" aria-labelledby="account-settings-title">
    <Button.Root href="/settings" variant="ghost" class="settings-back" onclick={(event) => navigate(event, '/settings')}><ArrowLeft /> Settings</Button.Root>
    <header class="settings-management-heading"><p class="settings-eyebrow">Account and app</p><h1 id="account-settings-title">Account</h1><p>{principal.email ?? 'Signed-in account'}</p></header>
    <dl class="settings-readonly-details"><div><dt>Profile editing</dt><dd>Not available</dd></div><div><dt>App</dt><dd>Stuff Stash web</dd></div></dl>
  </section>
{:else if !tenant || (route.settingsLevel === 'inventory' && !inventory)}
  <section class="workspace-main settings-management"><div class="settings-collection-state" role="alert"><h1>Settings unavailable</h1><p>The selected settings context is not available to this account.</p><Button.Root href="/settings" onclick={(event) => navigate(event, '/settings')}>Back to Settings</Button.Root></div></section>
{:else if !route.settingsCollection}
  <section class="workspace-main settings-management" aria-labelledby="settings-level-title">
    <Button.Root href="/settings" variant="ghost" class="settings-back" onclick={(event) => navigate(event, '/settings')}><ArrowLeft /> Settings</Button.Root>
    <header class="settings-management-heading"><p class="settings-eyebrow">{levelLabel}</p><h1 id="settings-level-title">{levelTitle}</h1>{#if inventory}<p>{inventory.name} belongs to {tenant.name}.</p>{:else}<p>Settings shared with this tenant’s inventories.</p>{/if}</header>
    <SettingsDestinationList label={`${levelTitle} settings`} destinations={route.settingsLevel === 'tenant' ? tenantSettingsDestinations(tenant) : inventorySettingsDestinations(inventory!)} {onNavigate} />
  </section>
{:else}
  <div class="workspace-main settings-management settings-management-resource">
    <Button.Root href={levelHref} variant="ghost" class="settings-back" onclick={(event) => navigate(event, levelHref)}><ArrowLeft /> {levelTitle}</Button.Root>
    {#if route.settingsCollection === 'tags' && inventory}
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
      <div class="settings-collection-state" role="alert"><h1>Section unavailable</h1><p>This settings section is not available in the selected context.</p></div>
    {/if}
  </div>
{/if}
