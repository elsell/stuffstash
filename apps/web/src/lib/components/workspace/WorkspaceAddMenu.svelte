<script lang="ts">
  import Plus from '@lucide/svelte/icons/plus';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { shellAddOptions, type ShellAddOption } from '$lib/application/workspaceShellNavigation';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import type { AssetKind } from '$lib/domain/inventory';
  import { cn } from '$lib/utils.js';

  let {
    tenantId,
    inventoryId,
    canOpen,
    disabledReason,
    disablePortal = false,
    onOpenAdd
  }: {
    tenantId: string | null;
    inventoryId: string | null;
    canOpen: boolean;
    disabledReason?: string;
    disablePortal?: boolean;
    onOpenAdd: (kind: AssetKind) => void;
  } = $props();

  let open = $state(false);
  const deniedNoteId = 'header-add-denied';

  function handleOpenChange(nextOpen: boolean): void {
    open = nextOpen;
  }

  function menuItemClass(primitiveClass: unknown): string {
    return cn(typeof primitiveClass === 'string' ? primitiveClass : undefined, 'add-menu-item');
  }

  function choose(event: MouseEvent, option: ShellAddOption, primitiveOnClick: unknown): void {
    if (shouldHandleWorkspaceLinkClick(event)) {
      event.preventDefault();
      open = false;
      onOpenAdd(option.kind);
    }
    if (typeof primitiveOnClick === 'function') primitiveOnClick(event);
  }
</script>

<div class="header-add-wrap">
  <DropdownMenu.Root bind:open onOpenChange={handleOpenChange}>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button.Root
          {...props}
          data-workspace-add-trigger="desktop"
          class="header-add min-h-11"
          disabled={!canOpen}
          aria-describedby={disabledReason ? deniedNoteId : undefined}
        >
          <Plus /> Add
        </Button.Root>
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content
      id="header-add-menu"
      class="add-menu"
      align="end"
      aria-label="Add asset kind"
      forceMount={disablePortal}
      portalProps={{ disabled: disablePortal }}
    >
      {#each shellAddOptions(tenantId, inventoryId) as option}
        <DropdownMenu.Item>
          {#snippet child({ props })}
            <a
              {...props}
              href={option.href}
              class={menuItemClass(props.class)}
              onclick={(event) => choose(event, option, props.onclick)}
            >{option.label}</a>
          {/snippet}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
  {#if disabledReason}
    <p id={deniedNoteId} class="visually-hidden" role="note">{disabledReason}</p>
  {/if}
</div>
