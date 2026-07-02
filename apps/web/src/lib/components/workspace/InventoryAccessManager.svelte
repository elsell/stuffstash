<script lang="ts">
  import Link2 from '@lucide/svelte/icons/link-2';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import UserPlus from '@lucide/svelte/icons/user-plus';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import {
    hasAccessPermission,
    type Inventory,
    type InventoryAccessGrant,
    type InventoryAccessInvitation,
    type InventoryAccessRelationship,
    type InvitationStatusFilter,
    type Tenant
  } from '$lib/domain/inventory';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenant,
    inventory,
    repository
  }: {
    tenant: Tenant | null;
    inventory: Inventory | null;
    repository: InventoryAccessRepository;
  } = $props();

  const relationships: InventoryAccessRelationship[] = ['viewer', 'editor'];
  const invitationStatuses: InvitationStatusFilter[] = ['all', 'pending', 'accepted', 'revoked', 'cancelled', 'expired'];
  const relationshipOptions = relationships.map((relationship) => ({
    value: relationship,
    label: relationship === 'viewer' ? 'Viewer' : 'Editor'
  }));
  const invitationStatusOptions = invitationStatuses.map((status) => ({
    value: status,
    label: status[0]?.toUpperCase() + status.slice(1)
  }));

  let grants = $state<InventoryAccessGrant[]>([]);
  let invitations = $state<InventoryAccessInvitation[]>([]);
  let grantNextCursor = $state<string | null>(null);
  let invitationNextCursor = $state<string | null>(null);
  let principalId = $state('');
  let grantRelationship = $state<InventoryAccessRelationship>('viewer');
  let invitationEmail = $state('');
  let invitationRelationship = $state<InventoryAccessRelationship>('viewer');
  let invitationStatus = $state<InvitationStatusFilter>('all');
  let inviteLinkToken = $state('');
  let busy = $state(false);
  let loaded = $state(false);
  let message = $state('');
  let error = $state('');
  let requestId = 0;

  let canShare = $derived(hasAccessPermission(inventory?.access, 'share'));
  let contextKey = $derived(tenant && inventory && canShare ? `${tenant.id}:${inventory.id}` : '');

  $effect(() => {
    requestId += 1;
    grants = [];
    invitations = [];
    grantNextCursor = null;
    invitationNextCursor = null;
    inviteLinkToken = '';
    invitationStatus = 'all';
    loaded = false;
    error = '';
    message = '';
    if (!contextKey) {
      return;
    }
    void loadAccess(contextKey, 'all');
  });

  async function loadAccess(expectedContext = contextKey, status = invitationStatus): Promise<void> {
    const context = snapshotContext(expectedContext);
    if (!context) {
      return;
    }
    const current = requestId;
    busy = true;
    error = '';
    try {
      const [grantPage, invitationPage] = await Promise.all([
        repository.listInventoryAccessGrants(context.tenantId, context.inventoryId),
        repository.listInventoryAccessInvitations(context.tenantId, context.inventoryId, status)
      ]);
      if (!sameContext(current, expectedContext)) {
        return;
      }
      grants = grantPage.items;
      invitations = invitationPage.items;
      grantNextCursor = grantPage.pagination.nextCursor;
      invitationNextCursor = invitationPage.pagination.nextCursor;
      loaded = true;
    } catch (caught) {
      if (sameContext(current, expectedContext)) {
        error = caught instanceof Error ? caught.message : 'Unable to load access.';
      }
    } finally {
      if (sameContext(current, expectedContext)) {
        busy = false;
      }
    }
  }

  async function loadMoreGrants(): Promise<void> {
    const context = snapshotContext(contextKey);
    if (!context || !grantNextCursor) {
      return;
    }
    const expectedContext = contextKey;
    const current = requestId;
    await mutate(expectedContext, async () => {
      const page = await repository.listInventoryAccessGrants(context.tenantId, context.inventoryId, grantNextCursor ?? undefined);
      if (!sameContext(current, expectedContext)) {
        return;
      }
      grants = [...grants, ...page.items];
      grantNextCursor = page.pagination.nextCursor;
    });
  }

  async function loadMoreInvitations(): Promise<void> {
    const context = snapshotContext(contextKey);
    if (!context || !invitationNextCursor) {
      return;
    }
    const expectedContext = contextKey;
    const current = requestId;
    await mutate(expectedContext, async () => {
      const page = await repository.listInventoryAccessInvitations(
        context.tenantId,
        context.inventoryId,
        invitationStatus,
        invitationNextCursor ?? undefined
      );
      if (!sameContext(current, expectedContext)) {
        return;
      }
      invitations = [...invitations, ...page.items];
      invitationNextCursor = page.pagination.nextCursor;
    });
  }

  async function loadInvitations(expectedContext: string, status: InvitationStatusFilter): Promise<void> {
    const context = snapshotContext(expectedContext);
    if (!context) {
      return;
    }
    const current = requestId;
    busy = true;
    error = '';
    try {
      const page = await repository.listInventoryAccessInvitations(context.tenantId, context.inventoryId, status);
      if (!sameContext(current, expectedContext)) {
        return;
      }
      invitations = page.items;
      invitationNextCursor = page.pagination.nextCursor;
      loaded = true;
    } catch (caught) {
      if (sameContext(current, expectedContext)) {
        error = caught instanceof Error ? caught.message : 'Unable to load invitations.';
      }
    } finally {
      if (sameContext(current, expectedContext)) {
        busy = false;
      }
    }
  }

  async function addGrant(): Promise<void> {
    const context = snapshotContext(contextKey);
    const targetPrincipalId = principalId.trim();
    const relationship = grantRelationship;
    if (!context || !targetPrincipalId) {
      return;
    }
    await mutate(context.key, async () => {
      const grant = await repository.grantInventoryAccess(context.tenantId, context.inventoryId, targetPrincipalId, relationship);
      if (!sameContext(context.requestId, context.key)) {
        return;
      }
      grants = [grant, ...grants.filter((candidate) => !sameGrant(candidate, grant))];
      principalId = '';
      message = `Granted ${grant.relationship} access.`;
    });
  }

  async function revokeGrant(grant: InventoryAccessGrant): Promise<void> {
    const context = snapshotContext(contextKey);
    if (!context) {
      return;
    }
    await mutate(context.key, async () => {
      await repository.revokeInventoryAccess(context.tenantId, context.inventoryId, grant.principalId, grant.relationship);
      if (!sameContext(context.requestId, context.key)) {
        return;
      }
      grants = grants.filter((candidate) => !sameGrant(candidate, grant));
      message = `Revoked ${grant.relationship} access.`;
    });
  }

  async function invite(): Promise<void> {
    const context = snapshotContext(contextKey);
    const email = invitationEmail.trim();
    const relationship = invitationRelationship;
    if (!context || !email) {
      return;
    }
    await mutate(context.key, async () => {
      const created = await repository.createInventoryAccessInvitation(context.tenantId, context.inventoryId, email, relationship);
      if (!sameContext(context.requestId, context.key)) {
        return;
      }
      invitations = [created.invitation, ...invitations.filter((candidate) => candidate.id !== created.invitation.id)];
      invitationEmail = '';
      inviteLinkToken = created.acceptanceToken ?? '';
      message = created.acceptanceToken ? 'Invitation created. Copy the one-time token now.' : 'Invitation created.';
    });
  }

  async function expireInvitation(invitation: InventoryAccessInvitation): Promise<void> {
    const context = snapshotContext(contextKey);
    if (!context) {
      return;
    }
    await mutate(context.key, async () => {
      const updated = await repository.updateInventoryAccessInvitationExpiration(
        context.tenantId,
        context.inventoryId,
        invitation.id,
        new Date(0).toISOString()
      );
      if (!sameContext(context.requestId, context.key)) {
        return;
      }
      invitations = replaceInvitation(updated);
      message = 'Invitation expiration updated.';
    });
  }

  async function cancelInvitation(invitation: InventoryAccessInvitation): Promise<void> {
    const context = snapshotContext(contextKey);
    if (!context) {
      return;
    }
    await mutate(context.key, async () => {
      await repository.cancelInventoryAccessInvitation(context.tenantId, context.inventoryId, invitation.id);
      if (!sameContext(context.requestId, context.key)) {
        return;
      }
      invitations = replaceInvitation({ ...invitation, status: 'cancelled' });
      message = 'Invitation cancelled.';
    });
  }

  async function deleteInvitation(invitation: InventoryAccessInvitation): Promise<void> {
    const context = snapshotContext(contextKey);
    if (!context) {
      return;
    }
    await mutate(context.key, async () => {
      await repository.deleteInventoryAccessInvitation(context.tenantId, context.inventoryId, invitation.id);
      if (!sameContext(context.requestId, context.key)) {
        return;
      }
      invitations = invitations.filter((candidate) => candidate.id !== invitation.id);
      message = 'Invitation deleted.';
    });
  }

  async function mutate(expectedContext: string, action: () => Promise<void>): Promise<void> {
    busy = true;
    error = '';
    message = '';
    try {
      await action();
    } catch (caught) {
      if (contextKey === expectedContext) {
        error = caught instanceof Error ? caught.message : 'Access action failed.';
      }
    } finally {
      if (contextKey === expectedContext) {
        busy = false;
      }
    }
  }

  function updateInvitationStatus(status: InvitationStatusFilter): void {
    invitationStatus = status;
    if (contextKey) {
      requestId += 1;
      invitations = [];
      invitationNextCursor = null;
      void loadInvitations(contextKey, status);
    }
  }

  function snapshotContext(expectedContext: string): { key: string; requestId: number; tenantId: string; inventoryId: string } | null {
    if (!tenant || !inventory || !canShare || !expectedContext) {
      return null;
    }
    return { key: expectedContext, requestId, tenantId: tenant.id, inventoryId: inventory.id };
  }

  function sameContext(expectedRequestId: number, expectedContext: string): boolean {
    return requestId === expectedRequestId && contextKey === expectedContext;
  }

  function replaceInvitation(invitation: InventoryAccessInvitation): InventoryAccessInvitation[] {
    return invitations.map((candidate) => (candidate.id === invitation.id ? invitation : candidate));
  }

  function sameGrant(left: InventoryAccessGrant, right: InventoryAccessGrant): boolean {
    return (
      left.tenantId === right.tenantId &&
      left.inventoryId === right.inventoryId &&
      left.principalId === right.principalId &&
      left.relationship === right.relationship
    );
  }
</script>

<section class="settings-panel wide" aria-labelledby="settings-access">
  <div class="settings-panel-heading">
    <UserPlus aria-hidden="true" />
    <div>
      <h2 id="settings-access">Sharing</h2>
      <p>{canShare ? 'Manage direct grants and invite links for this inventory.' : 'Sharing requires inventory share access.'}</p>
    </div>
  </div>

  {#if !inventory}
    <p class="denied-note">Select an inventory before managing sharing.</p>
  {:else if !canShare}
    <p class="denied-note" role="alert">You can view this inventory, but you cannot manage sharing.</p>
  {:else}
    {#if error}
      <p class="denied-note" role="alert">{error}</p>
    {/if}
    {#if message}
      <p class="success-note" role="status">{message}</p>
    {/if}
    {#if inviteLinkToken}
      <p class="token-line one-time-token" role="status"><Link2 aria-hidden="true" /> {inviteLinkToken}</p>
    {/if}

    <form class="access-form" onsubmit={(event) => { event.preventDefault(); void addGrant(); }}>
      <div class="field-stack">
        <Label for="grant-principal">Principal ID</Label>
        <Input id="grant-principal" bind:value={principalId} placeholder="principal-id" />
      </div>
      <div class="field-stack">
        <span class="field-label">Grant relationship</span>
        <SegmentedControl
          label="Grant relationship"
          value={grantRelationship}
          options={relationshipOptions}
          onSelect={(value) => { grantRelationship = value as InventoryAccessRelationship; }}
        />
      </div>
      <Button.Root type="submit" disabled={busy || principalId.trim().length === 0}>Grant access</Button.Root>
    </form>

    <form class="access-form" onsubmit={(event) => { event.preventDefault(); void invite(); }}>
      <div class="field-stack">
        <Label for="invite-email">Invitee email</Label>
        <Input id="invite-email" type="email" bind:value={invitationEmail} placeholder="person@example.com" />
      </div>
      <div class="field-stack">
        <span class="field-label">Invitation relationship</span>
        <SegmentedControl
          label="Invitation relationship"
          value={invitationRelationship}
          options={relationshipOptions}
          onSelect={(value) => { invitationRelationship = value as InventoryAccessRelationship; }}
        />
      </div>
      <Button.Root type="submit" disabled={busy || invitationEmail.trim().length === 0}>Create invite</Button.Root>
    </form>

    <div class="access-columns">
      <div class="access-list" aria-label="Direct grants">
        <h3>Direct grants</h3>
        {#if busy && !loaded}
          <p class="muted-note">Loading grants...</p>
        {:else if grants.length === 0}
          <p class="muted-note">No direct grants.</p>
        {:else}
          {#each grants as grant}
            <div class="access-row">
              <span>
                <strong>{grant.principalId}</strong>
                <small>{grant.relationship}</small>
              </span>
              <Button.Root variant="outline" size="sm" disabled={busy} onclick={() => { void revokeGrant(grant); }}>Revoke</Button.Root>
            </div>
          {/each}
          {#if grantNextCursor}
            <Button.Root variant="outline" size="sm" disabled={busy} onclick={() => { void loadMoreGrants(); }}>Load more grants</Button.Root>
          {/if}
        {/if}
      </div>

      <div class="access-list" aria-label="Invitations">
        <div class="access-list-header">
          <h3>Invitations</h3>
          <SegmentedControl
            label="Invitation status"
            value={invitationStatus}
            options={invitationStatusOptions}
            onSelect={(value) => updateInvitationStatus(value as InvitationStatusFilter)}
          />
        </div>
        {#if busy && !loaded}
          <p class="muted-note">Loading invitations...</p>
        {:else if invitations.length === 0}
          <p class="muted-note">No invitations.</p>
        {:else}
          {#each invitations as invitation}
            <div class="access-row invitation-row">
              <span>
                <strong>{invitation.email}</strong>
                <small>{invitation.relationship} / {invitation.status}{invitation.isExpired ? ' / expired' : ''}</small>
              </span>
              <Badge variant={invitation.status === 'pending' && !invitation.isExpired ? 'secondary' : 'outline'}>
                {invitation.status}
              </Badge>
              <div class="access-actions">
                <Button.Root variant="outline" size="sm" disabled={busy || invitation.status !== 'pending'} onclick={() => { void expireInvitation(invitation); }}>Expire</Button.Root>
                <Button.Root variant="outline" size="sm" disabled={busy || invitation.status !== 'pending'} onclick={() => { void cancelInvitation(invitation); }}>Cancel</Button.Root>
                <Button.Root variant="ghost" size="icon-sm" disabled={busy} aria-label={`Delete invitation for ${invitation.email}`} onclick={() => { void deleteInvitation(invitation); }}><Trash2 /></Button.Root>
              </div>
            </div>
          {/each}
          {#if invitationNextCursor}
            <Button.Root variant="outline" size="sm" disabled={busy} onclick={() => { void loadMoreInvitations(); }}>Load more invitations</Button.Root>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
</section>
