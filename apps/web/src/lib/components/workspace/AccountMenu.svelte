<script lang="ts">
  import ChevronUp from '@lucide/svelte/icons/chevron-up';
  import LogOut from '@lucide/svelte/icons/log-out';
  import Settings from '@lucide/svelte/icons/settings';
  import UserRound from '@lucide/svelte/icons/user-round';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { cn } from '$lib/utils.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Sheet from '$lib/components/ui/sheet/index.js';

  let {
    userLabel,
    settingsHref,
    mobile = false,
    disablePortal = false,
    onSignOut,
    onOpenSettings,
    onOpenChange: notifyOpenChange
  }: {
    userLabel: string;
    settingsHref: string;
    mobile?: boolean;
    disablePortal?: boolean;
    onSignOut: () => void;
    onOpenSettings: () => void;
    onOpenChange?: (open: boolean) => void;
  } = $props();

  let open = $state(false);

  function handleOpenChange(nextOpen: boolean): void {
    open = nextOpen;
    notifyOpenChange?.(nextOpen);
  }

  function signOut(): void {
    handleOpenChange(false);
    onSignOut();
  }

  function openSettings(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    handleOpenChange(false);
    onOpenSettings();
  }

  function accountMenuLinkClass(primitiveClass: unknown): string {
    return cn(typeof primitiveClass === 'string' ? primitiveClass : undefined, 'account-menu-link');
  }

  function openSettingsFromMenu(event: MouseEvent, primitiveOnClick: unknown): void {
    openSettings(event);
    if (typeof primitiveOnClick === 'function') primitiveOnClick(event);
  }
</script>

{#snippet identity()}
  <span class="account-icon" aria-hidden="true"><UserRound /></span>
  <span class="account-copy">
    <strong>Account</strong>
    <small>{userLabel}</small>
  </span>
{/snippet}

{#if mobile}
  <div class="mobile-account-menu">
    <Sheet.Root bind:open onOpenChange={handleOpenChange}>
      <Sheet.Trigger>
        {#snippet child({ props })}
          <Button.Root
            {...props}
            variant="ghost"
            size="icon"
            class="mobile-account-trigger"
            style="min-width: 44px; min-height: 44px"
            aria-label="Open account menu"
          ><UserRound /></Button.Root>
        {/snippet}
      </Sheet.Trigger>
      {#if open || !disablePortal}
        <Sheet.Content
          side="bottom"
          forceMount={disablePortal}
          class="account-sheet rounded-t-2xl"
          portalProps={{ disabled: disablePortal }}
        >
          <Sheet.Header class="account-sheet-header">
            <Sheet.Title>Account</Sheet.Title>
            <Sheet.Description>Signed in to Stuff Stash</Sheet.Description>
          </Sheet.Header>
          <div class="account-sheet-identity">
            {@render identity()}
          </div>
          <div class="account-sheet-navigation">
            <Button.Root href={settingsHref} variant="outline" class="account-settings" onclick={openSettings}><Settings /> Settings</Button.Root>
          </div>
          <Sheet.Footer class="account-sheet-footer">
            <Button.Root data-variant="destructive" variant="destructive" class="account-sign-out" onclick={signOut}><LogOut /> Sign out</Button.Root>
          </Sheet.Footer>
        </Sheet.Content>
      {/if}
    </Sheet.Root>
  </div>
{:else}
  <DropdownMenu.Root bind:open onOpenChange={handleOpenChange}>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button.Root
          {...props}
          variant="ghost"
          class="account-trigger"
          aria-label={`Account menu for ${userLabel}`}
        >
          {@render identity()}
          <ChevronUp class="account-chevron" aria-hidden="true" />
        </Button.Root>
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content
      side="right"
      align="end"
      sideOffset={8}
      forceMount={disablePortal}
      portalProps={{ disabled: disablePortal }}
      class="account-dropdown w-64"
    >
      <DropdownMenu.Label class="account-dropdown-identity">
        <span>Signed in as</span>
        <strong>{userLabel}</strong>
      </DropdownMenu.Label>
      <DropdownMenu.Separator />
      <DropdownMenu.Item>
        {#snippet child({ props })}
          <a
            {...props}
            class={accountMenuLinkClass(props.class)}
            href={settingsHref}
            onclick={(event) => openSettingsFromMenu(event, props.onclick)}
          ><Settings /> Settings</a>
        {/snippet}
      </DropdownMenu.Item>
      <DropdownMenu.Separator />
      <DropdownMenu.Item variant="destructive" class="min-h-11" onclick={signOut}><LogOut /> Sign out</DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<style>
  :global(.account-trigger) {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    width: 100%;
    height: auto;
    min-height: 56px;
    gap: var(--space-3);
    justify-content: stretch;
    padding: var(--space-2);
    text-align: left;
  }

  .account-icon {
    display: inline-flex;
    width: 32px;
    height: 32px;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--border);
    border-radius: var(--radius-pill);
    background: var(--muted);
    color: var(--foreground);
  }

  .account-icon :global(svg) {
    width: 17px;
    height: 17px;
  }

  .account-copy,
  :global(.account-dropdown-identity) {
    display: grid;
    min-width: 0;
    gap: 2px;
  }

  .account-copy strong,
  .account-copy small,
  :global(.account-dropdown-identity strong) {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .account-copy strong {
    font-weight: 600;
  }

  .account-copy small {
    color: var(--muted-foreground);
    font-size: var(--text-caption-size);
    line-height: var(--text-caption-line-height);
  }

  :global(.account-chevron) {
    color: var(--muted-foreground);
  }

  :global(.account-dropdown-identity strong) {
    color: var(--foreground);
    font-size: var(--text-metadata-size);
    line-height: var(--text-metadata-line-height);
  }

  .mobile-account-menu {
    display: none;
  }

  :global(.mobile-account-trigger) {
    width: 44px;
    height: 44px;
    min-width: 44px;
    min-height: 44px;
  }

  :global(.account-sheet) {
    max-height: min(80dvh, 32rem);
  }

  :global(.account-sheet-header) {
    padding-right: 64px;
  }

  :global(.account-sheet-identity) {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    gap: var(--space-3);
    margin-inline: var(--space-4);
    padding: var(--space-4);
    border: 1px solid var(--border);
    border-radius: var(--radius-surface);
    background: var(--card);
  }

  :global(.account-sheet-navigation) {
    margin-inline: var(--space-4);
  }

  :global(.account-settings),
  :global(.account-menu-link) {
    width: 100%;
    min-height: 44px;
    justify-content: flex-start;
  }

  :global(.account-sheet-footer) {
    margin-top: var(--space-4);
    border-top: 1px solid var(--border);
    padding-bottom: max(var(--space-4), env(safe-area-inset-bottom));
  }

  :global(.account-sign-out) {
    min-height: 44px;
  }

  @media (max-width: 900px) {
    .mobile-account-menu {
      display: block;
    }
  }
</style>
