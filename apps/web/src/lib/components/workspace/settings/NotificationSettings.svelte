<script lang="ts">
  import { onMount, untrack } from 'svelte';
  import type { CustomAssetType } from '$lib/domain/inventory';
  import type { NotificationPreferences } from '$lib/domain/notification';
  import type { NotificationRepository } from '$lib/ports/notificationRepository';
  import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
  import { NotificationPreferencesSession } from '$lib/application/notificationPreferencesSession';
  import { safeWorkspaceErrorMessage } from '$lib/application/workspaceSafeError';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import ExpirationReminderEditor from './ExpirationReminderEditor.svelte';

  let { tenantId, inventoryId, initialTimezone, repository, observer, assetTypes }: {
    tenantId: string; inventoryId: string; initialTimezone: string; repository: NotificationRepository;
    observer: WorkspaceObserver; assetTypes: CustomAssetType[];
  } = $props();
  const session = untrack(() => new NotificationPreferencesSession(repository, observer, tenantId, inventoryId));
  const timezoneId = $props.id();
  let preferences = $state<NotificationPreferences | null>(null);
  let busy = $state(false);
  let error = $state('');
  let timezone = $state('');
  let timezoneSaved = $state(false);
  let validTimezone = $derived.by(() => {
    try { new Intl.DateTimeFormat('en', { timeZone: timezone }); return timezone.trim().length > 0; }
    catch { return false; }
  });
  let types = $derived(assetTypes.filter((type) => type.lifecycleState === 'active' && type.expirationEnabled));
  onMount(() => { void load(); });

  async function load() {
    if (busy) return;
    busy = true; error = '';
    try {
      const first = preferences === null;
      preferences = first ? await session.initialize(initialTimezone) : await session.refresh();
      if (first) timezone = preferences.timezone;
    } catch (caught) { error = safeWorkspaceErrorMessage(caught, 'Reminder settings could not be loaded. Try again.'); }
    finally { busy = false; }
  }
  async function save(operation: () => Promise<NotificationPreferences>) {
    if (busy) throw new Error('Another settings request is in progress.');
    busy = true; error = '';
    try { preferences = await operation(); }
    catch (caught) {
      error = 'Your changes are still here. If settings changed on another device, refresh saved settings before saving again.';
      throw caught;
    } finally { busy = false; }
  }
  async function saveTimezone(event: SubmitEvent) {
    event.preventDefault();
    if (!validTimezone || busy) return;
    timezoneSaved = false;
    try { await save(() => session.saveTimezone(timezone)); timezoneSaved = true; }
    catch (caught) { error = safeWorkspaceErrorMessage(caught, 'Timezone could not be saved. Refresh saved settings and try again.'); }
  }
</script>

<section aria-labelledby={`${timezoneId}-title`}>
  <header><h1 id={`${timezoneId}-title`}>Notifications</h1><p>Your reminders for this inventory. Other members have their own settings.</p></header>
  {#if error}<p role="alert">{error}</p>{/if}
  {#if !preferences}
    {#if busy}<p role="status">Loading reminders…</p>{:else}<Button.Root onclick={load}>Retry loading reminders</Button.Root>{/if}
  {:else}
    <Button.Root variant="outline" disabled={busy} onclick={load}>Refresh saved settings</Button.Root>
    <fieldset disabled={busy}>
      <legend class="sr-only">Personal notification settings</legend>
      <section aria-label="Inventory defaults"><h2>Inventory defaults</h2>
        <ExpirationReminderEditor initialPolicy={preferences.defaults} onSave={async (policy) => { if (policy) await save(() => session.saveDefaults(policy)); }} />
      </section>
      <section aria-label="Calendar timezone"><h2>Calendar timezone</h2>
        <form onsubmit={saveTimezone}>
          <Label for={timezoneId}>Timezone</Label>
          <Input id={timezoneId} value={timezone} oninput={(event) => { timezone = event.currentTarget.value; timezoneSaved = false; }} aria-invalid={!validTimezone} aria-describedby={`${timezoneId}-help`} />
          <p id={`${timezoneId}-help`}>{validTimezone ? `Saved timezone: ${preferences.timezone}. Dates end at midnight in this timezone.` : 'Enter a timezone such as America/New_York or Europe/London.'}</p>
          <Button.Root type="submit" disabled={!validTimezone}>Save timezone</Button.Root>
          {#if timezoneSaved}<p role="status">Timezone saved.</p>{/if}
        </form>
      </section>
      <section aria-label="Asset type reminders"><h2>Asset type reminders</h2>
        {#each types as type (type.id)}
          <details><summary>{type.displayName} · {preferences.overrides.some((value) => value.customAssetTypeId === type.id) ? 'Custom' : 'Inherited'}</summary>
            <ExpirationReminderEditor initialPolicy={preferences.overrides.find((value) => value.customAssetTypeId === type.id)?.settings ?? null} inheritedPolicy={preferences.defaults} onSave={(policy) => save(() => session.saveTypeOverride(type.id, policy))} />
          </details>
        {:else}<p>Enable expiration tracking on an asset type to customize its reminders here.</p>{/each}
      </section>
    </fieldset>
    <p>Set up push notifications in the mobile app. Your inbox is available on web and mobile.</p>
  {/if}
</section>

<style>
  section, fieldset, form { display: grid; gap: 1rem; }
  fieldset { min-width: 0; border: 0; padding: 0; }
  fieldset > section { padding-block: 1rem; border-bottom: 1px solid var(--border); }
  h1 { font-size: 1.5rem; font-weight: 650; }
  h2 { font-size: 1.125rem; font-weight: 600; }
  p { color: var(--muted-foreground); }
  summary { cursor: pointer; padding-block: 0.75rem; min-height: 2.75rem; }
  [role='alert'] { color: var(--destructive); }
</style>
