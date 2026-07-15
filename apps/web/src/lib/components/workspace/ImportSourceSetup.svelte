<script lang="ts">
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import FileText from '@lucide/svelte/icons/file-text';
  import Image from '@lucide/svelte/icons/image';
  import LockKeyhole from '@lucide/svelte/icons/lock-keyhole';
  import Server from '@lucide/svelte/icons/server';
  import type { Inventory } from '$lib/domain/inventory';
  import type { ImportCSVSelection, ImportSourceChoice } from '$lib/application/workspaceImportRequest';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Checkbox from '$lib/components/ui/checkbox/index.js';
  import * as Input from '$lib/components/ui/input/index.js';
  import * as Label from '$lib/components/ui/label/index.js';
  import ImportFlowStepper from './ImportFlowStepper.svelte';

  type Props = {
    sourceChoice: ImportSourceChoice;
    availableSteps: Array<'source' | 'connect' | 'preview' | 'run'>;
    baseUrl: string;
    username: string;
    password: string;
    includeImages: boolean;
    allowPrivateNetwork: boolean;
    allowInsecureTLS: boolean;
    csvSelection: ImportCSVSelection | null;
    canConfirmSource: boolean;
    busy: boolean;
    inventory: Inventory | null;
    onFileSelected: (event: Event) => void;
    onConfirmSource: () => void;
    onNavigateStep: (step: 'source' | 'connect' | 'preview' | 'run') => void;
    onBack: () => void;
  };

  let {
    sourceChoice,
    availableSteps,
    baseUrl = $bindable(''),
    username = $bindable(''),
    password = $bindable(''),
    includeImages = $bindable(true),
    allowPrivateNetwork = $bindable(false),
    allowInsecureTLS = $bindable(false),
    csvSelection,
    canConfirmSource,
    busy,
    inventory,
    onFileSelected,
    onConfirmSource,
    onNavigateStep,
    onBack
  }: Props = $props();
</script>

<Card.Root>
  <Card.Header>
    <ImportFlowStepper current="connect" {availableSteps} {onNavigateStep} />
    <Card.Title>{sourceChoice === 'homebox_live' ? 'Connect to Homebox' : 'Upload Homebox CSV'}</Card.Title>
    <Card.Description>Stuff Stash will verify the source and build a preview.</Card.Description>
  </Card.Header>
  <Card.Content class="import-source-setup-content">
    {#if sourceChoice === 'homebox_live'}
      <div class="setup-grid">
        <div class="setup-fields">
          <div class="field-stack">
            <Label.Root for="homebox-url">Homebox URL</Label.Root>
            <Input.Root
              id="homebox-url"
              bind:value={baseUrl}
              placeholder="homebox.example.com or http://homebox.local:3100"
              autocomplete="url"
              autocapitalize="none"
              autocorrect="off"
              inputmode="url"
              spellcheck={false}
            />
            <small class="field-note">Explicit http:// and https:// URLs are preserved. Schemeless hosts try https:// first.</small>
          </div>
          <div class="field-grid">
            <div class="field-stack">
              <Label.Root for="homebox-user">Email</Label.Root>
              <Input.Root
                id="homebox-user"
                type="email"
                bind:value={username}
                autocomplete="username"
                autocapitalize="none"
                autocorrect="off"
                inputmode="email"
                spellcheck={false}
              />
            </div>
            <div class="field-stack">
              <Label.Root for="homebox-password">Password</Label.Root>
              <Input.Root
                id="homebox-password"
                type="password"
                bind:value={password}
                autocomplete="current-password"
                autocapitalize="none"
                autocorrect="off"
                spellcheck={false}
              />
            </div>
          </div>
          <Label.Root class="import-check-row">
            <Checkbox.Root bind:checked={includeImages} />
            <span>Import photos when Homebox provides them</span>
          </Label.Root>
          <details class="advanced-options">
            <summary>Connection options</summary>
            <div class="advanced-option-list">
              <Label.Root class="import-check-row">
                <Checkbox.Root bind:checked={allowPrivateNetwork} />
                <span>Allow private-network Homebox URL</span>
              </Label.Root>
              <Label.Root class="import-check-row">
                <Checkbox.Root bind:checked={allowInsecureTLS} />
                <span>Allow self-signed TLS certificate</span>
              </Label.Root>
            </div>
          </details>
        </div>
        <div class="connection-summary" aria-label="Connection summary">
          <span><Server size={16} aria-hidden="true" />Live Homebox API</span>
          <span><LockKeyhole size={16} aria-hidden="true" />Credentials are encrypted for preview and import</span>
          <span><Image size={16} aria-hidden="true" />{includeImages ? 'Photos will be imported' : 'Photos will be skipped'}</span>
        </div>
      </div>
    {:else}
      <div class="setup-grid">
        <div class="field-stack">
          <Label.Root for="homebox-csv">Homebox CSV export</Label.Root>
          <Input.Root
            id="homebox-csv"
            type="file"
            accept=".csv,text/csv"
            autocapitalize="none"
            autocorrect="off"
            spellcheck={false}
            onchange={onFileSelected}
          />
          {#if csvSelection}
            <small class="field-note">{csvSelection.name} · {Math.max(1, Math.round(csvSelection.size / 1024))} KB · photos are not included in CSV exports</small>
          {:else}
            <small class="field-note">CSV files must be 10 MiB or smaller. Homebox CSV exports do not include photos.</small>
          {/if}
        </div>
        <div class="connection-summary" aria-label="CSV import summary">
          <span><FileText size={16} aria-hidden="true" />CSV snapshot</span>
          <span><Image size={16} aria-hidden="true" />Photos unavailable</span>
        </div>
      </div>
    {/if}

    <div class="action-row">
      <Button.Root onclick={onConfirmSource} disabled={!canConfirmSource || busy || !inventory}>
        <Button.BusyContent
          {busy}
          icon={CheckCircle2}
          label={sourceChoice === 'homebox_live' ? 'Confirm connection' : 'Prepare preview'}
          busyLabel={sourceChoice === 'homebox_live' ? 'Confirming connection' : 'Preparing preview'}
        />
      </Button.Root>
      <Button.Root variant="outline" onclick={onBack} disabled={busy}>Back</Button.Root>
    </div>
  </Card.Content>
</Card.Root>

<style>
  :global(.import-source-setup-content) {
    display: grid;
    gap: 1rem;
  }

  .setup-grid {
    align-items: start;
    display: grid;
    gap: 1rem;
    grid-template-columns: minmax(0, 1fr) minmax(13rem, 0.34fr);
  }

  .setup-fields {
    display: grid;
    gap: 1rem;
    min-width: 0;
  }

  .field-grid {
    display: grid;
    gap: 0.75rem;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .field-stack {
    display: grid;
    gap: 0.35rem;
    min-width: 0;
  }

  .field-note {
    color: var(--muted-foreground);
    font-size: 0.82rem;
  }

  .action-row,
  :global(.import-check-row) {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  .action-row {
    flex-wrap: wrap;
    scroll-margin-bottom: var(--mobile-scroll-clearance, 10rem);
  }

  .connection-summary {
    border-left: 1px solid var(--border);
    display: grid;
    gap: 0.55rem;
    padding-left: 0.75rem;
  }

  .connection-summary span {
    align-items: center;
    color: var(--muted-foreground);
    display: flex;
    font-size: 0.82rem;
    gap: 0.5rem;
    min-width: 0;
  }

  .advanced-options {
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 0.75rem;
  }

  .advanced-options summary {
    color: var(--muted-foreground);
    cursor: pointer;
    font-size: 0.88rem;
    font-weight: 600;
  }

  .advanced-options[open] summary {
    color: var(--foreground);
    margin-bottom: 0.7rem;
  }

  .advanced-option-list {
    display: grid;
    gap: 0.6rem;
  }

  @media (max-width: 860px) {
    .setup-grid {
      grid-template-columns: 1fr;
    }

    :global(.import-check-row) {
      align-items: flex-start;
    }

    .connection-summary {
      border-left: 0;
      border-top: 1px solid var(--border);
      padding-left: 0;
      padding-top: 0.75rem;
    }

    .field-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
