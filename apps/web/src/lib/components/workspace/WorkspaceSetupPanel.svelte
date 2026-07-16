<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import {
    validateWorkspaceSetupDraft,
    workspaceSetupDescription,
    workspaceSetupTitle,
    type WorkspaceSetupMode
  } from '$lib/application/workspaceOnboarding';

  let {
    mode,
    tenantName = '',
    busy = false,
    error = '',
    submitLabel = 'Create workspace',
    onSubmit
  }: {
    mode: WorkspaceSetupMode;
    tenantName?: string;
    busy?: boolean;
    error?: string;
    submitLabel?: string;
    onSubmit: (draft: { tenantName: string; inventoryName: string }) => Promise<void>;
  } = $props();

  let tenantDraft = $state('');
  let inventoryDraft = $state('');
  let tenantError = $state('');
  let inventoryError = $state('');
  let formError = $derived(error);
  let showTenantField = $derived(mode === 'tenant_and_inventory');

  async function submit(): Promise<void> {
    const validation = validateWorkspaceSetupDraft(mode, { tenantName: tenantDraft, inventoryName: inventoryDraft });
    tenantError = validation.tenantError;
    inventoryError = validation.inventoryError;
    if (!validation.valid) {
      return;
    }
    await onSubmit({ tenantName: validation.tenantName, inventoryName: validation.inventoryName });
  }
</script>

<Card.Root class="setup-panel">
  <Card.Header>
    <Card.Title>{workspaceSetupTitle(mode)}</Card.Title>
    <Card.Description>{workspaceSetupDescription(mode, tenantName)}</Card.Description>
  </Card.Header>
  <Card.Content>
    <form class="setup-form" onsubmit={(event) => { event.preventDefault(); void submit(); }}>
      {#if showTenantField}
        <div class="field-stack">
          <Label for="workspace-tenant-name">Tenant name</Label>
          <Input
            id="workspace-tenant-name"
            bind:value={tenantDraft}
            autocomplete="organization"
            aria-invalid={tenantError ? 'true' : undefined}
            aria-describedby={tenantError ? 'workspace-tenant-error' : undefined}
          />
          {#if tenantError}<p id="workspace-tenant-error" class="field-error">{tenantError}</p>{/if}
        </div>
      {/if}

      <div class="field-stack">
        <Label for="workspace-inventory-name">Inventory name</Label>
        <Input
          id="workspace-inventory-name"
          bind:value={inventoryDraft}
          autocomplete="off"
          aria-invalid={inventoryError ? 'true' : undefined}
          aria-describedby={inventoryError ? 'workspace-inventory-error' : undefined}
        />
        {#if inventoryError}<p id="workspace-inventory-error" class="field-error">{inventoryError}</p>{/if}
      </div>

      {#if formError}
        <p class="form-error" role="alert">{formError}</p>
      {/if}

      <Button.Root type="submit" disabled={busy}>{busy ? 'Creating...' : submitLabel}</Button.Root>
    </form>
  </Card.Content>
</Card.Root>
