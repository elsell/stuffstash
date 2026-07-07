<script lang="ts">
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import type { ImportJobPreview } from '$lib/domain/inventory';
  import * as Button from '$lib/components/ui/button/index.js';
  import { fileSizeLabel, previewAssetContext, previewLocationContext } from './importWorkspacePresentation';

  const PLAN_PAGE_SIZE = 8;

  type PlanSectionId = 'fields' | 'locations' | 'assets' | 'attachments';

  type PlanColumn = {
    key: string;
    label: string;
  };

  type PlanRow = {
    id: string;
    cells: Record<string, string>;
  };

  type PlanSection = {
    id: PlanSectionId;
    title: string;
    emptyText: string;
    truncated: boolean;
    columns: PlanColumn[];
    rows: PlanRow[];
  };

  type Props = {
    preview: ImportJobPreview;
  };

  let { preview }: Props = $props();
  let pageBySection = $state<Record<PlanSectionId, number>>({
    fields: 0,
    locations: 0,
    assets: 0,
    attachments: 0
  });

  let sections = $derived(buildSections(preview));

  $effect(() => {
    for (const section of sections) {
      const pageCount = planPageCount(section);
      if (pageBySection[section.id] >= pageCount) {
        pageBySection = {
          ...pageBySection,
          [section.id]: Math.max(0, pageCount - 1)
        };
      }
    }
  });

  function buildSections(preview: ImportJobPreview): PlanSection[] {
    return [
      {
        id: 'fields',
        title: 'Fields',
        emptyText: 'No custom fields planned.',
        truncated: preview.fieldsTruncated,
        columns: [
          { key: 'name', label: 'Field' },
          { key: 'key', label: 'Key' },
          { key: 'type', label: 'Type' }
        ],
        rows: preview.fields.map((field) => ({
          id: field.key,
          cells: {
            name: field.displayName || field.key,
            key: field.key,
            type: field.type
          }
        }))
      },
      {
        id: 'locations',
        title: 'Locations',
        emptyText: 'No locations planned.',
        truncated: preview.locationsTruncated,
        columns: [
          { key: 'name', label: 'Location' },
          { key: 'kind', label: 'Kind' },
          { key: 'context', label: 'Context' }
        ],
        rows: preview.locations.map((item, index) => ({
          id: `location-${index}-${item.title}`,
          cells: {
            name: item.title,
            kind: item.kind,
            context: previewLocationContext(item)
          }
        }))
      },
      {
        id: 'assets',
        title: 'Assets',
        emptyText: 'No asset records planned.',
        truncated: preview.assetsTruncated,
        columns: [
          { key: 'name', label: 'Asset' },
          { key: 'kind', label: 'Kind' },
          { key: 'context', label: 'Context' }
        ],
        rows: preview.assets.map((item, index) => ({
          id: `asset-${index}-${item.title}`,
          cells: {
            name: item.title,
            kind: item.kind,
            context: previewAssetContext(item)
          }
        }))
      },
      {
        id: 'attachments',
        title: 'Photos/files',
        emptyText: 'No photos or files planned.',
        truncated: preview.attachmentsTruncated,
        columns: [
          { key: 'name', label: 'File' },
          { key: 'type', label: 'Type' },
          { key: 'size', label: 'Size' }
        ],
        rows: preview.attachments.map((attachment, index) => ({
          id: `attachment-${index}-${attachment.fileName || 'unnamed'}`,
          cells: {
            name: `${attachment.fileName || 'Unnamed attachment'}${attachment.primary ? ' (primary)' : ''}`,
            type: attachment.contentType || 'unknown type',
            size: fileSizeLabel(attachment.sizeBytes)
          }
        }))
      }
    ];
  }

  function planPageCount(section: PlanSection): number {
    return Math.max(1, Math.ceil(section.rows.length / PLAN_PAGE_SIZE));
  }

  function visibleRows(section: PlanSection): PlanRow[] {
    const start = visibleStart(section);
    return section.rows.slice(start, start + PLAN_PAGE_SIZE);
  }

  function visibleStart(section: PlanSection): number {
    return Math.min(pageBySection[section.id] * PLAN_PAGE_SIZE, Math.max(0, section.rows.length - 1));
  }

  function visibleEnd(section: PlanSection): number {
    return Math.min(section.rows.length, visibleStart(section) + visibleRows(section).length);
  }

  function sectionCountLabel(section: PlanSection): string {
    if (section.rows.length === 0) return 'None planned';
    if (section.rows.length > PLAN_PAGE_SIZE || section.truncated) {
      return `${visibleStart(section) + 1}-${visibleEnd(section)} of ${section.rows.length}${section.truncated ? '+' : ''}`;
    }
    return `${section.rows.length} ${section.rows.length === 1 ? 'record' : 'records'}`;
  }

  function setPage(section: PlanSection, nextPage: number): void {
    pageBySection = {
      ...pageBySection,
      [section.id]: Math.max(0, Math.min(planPageCount(section) - 1, nextPage))
    };
  }
</script>

<div class="preview-plan-sections">
  {#each sections as section}
    <section class="plan-section" aria-labelledby={`import-plan-${section.id}`}>
      <div class="plan-section-heading">
        <div>
          <h3 id={`import-plan-${section.id}`}>{section.title}</h3>
          <small>{sectionCountLabel(section)}</small>
        </div>
        {#if section.truncated}
          <span class="partial-list-badge">Partial list</span>
        {/if}
      </div>

      {#if section.rows.length === 0}
        <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> {section.emptyText}</div>
      {:else}
        <div class="plan-table-wrap">
          <table class="plan-table" aria-label={`${section.title} plan preview`}>
            <thead>
              <tr>
                {#each section.columns as column}
                  <th scope="col">{column.label}</th>
                {/each}
              </tr>
            </thead>
            <tbody>
              {#each visibleRows(section) as row}
                <tr>
                  {#each section.columns as column, index}
                    <td class:primary-cell={index === 0}>{row.cells[column.key]}</td>
                  {/each}
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
        {#if section.rows.length > PLAN_PAGE_SIZE}
          <div class="plan-pagination">
            <span>Page {pageBySection[section.id] + 1} of {planPageCount(section)}</span>
            <Button.Root
              variant="outline"
              size="icon"
              disabled={pageBySection[section.id] === 0}
              aria-label={`Previous ${section.title.toLowerCase()} plan page`}
              onclick={() => setPage(section, pageBySection[section.id] - 1)}
            >
              <ChevronLeft size={16} aria-hidden="true" />
            </Button.Root>
            <Button.Root
              variant="outline"
              size="icon"
              disabled={pageBySection[section.id] >= planPageCount(section) - 1}
              aria-label={`Next ${section.title.toLowerCase()} plan page`}
              onclick={() => setPage(section, pageBySection[section.id] + 1)}
            >
              <ChevronRight size={16} aria-hidden="true" />
            </Button.Root>
          </div>
        {/if}
      {/if}
    </section>
  {/each}
</div>

<style>
  .preview-plan-sections {
    display: grid;
    gap: 0.75rem;
  }

  .plan-section {
    border-top: 1px solid var(--border);
    display: grid;
    gap: 0.65rem;
    min-width: 0;
    padding-top: 0.75rem;
  }

  .plan-section:first-child {
    border-top: 0;
    padding-top: 0;
  }

  .plan-section-heading,
  .plan-pagination,
  .quiet-row {
    align-items: center;
    display: flex;
    gap: 0.5rem;
  }

  .plan-section-heading {
    justify-content: space-between;
  }

  .plan-section-heading > div {
    min-width: 0;
  }

  .plan-section-heading small {
    color: var(--muted-foreground);
    display: block;
    font-size: 0.78rem;
    margin-top: 0.1rem;
  }

  .partial-list-badge {
    background: color-mix(in oklab, var(--muted) 54%, transparent);
    border: 1px solid var(--border);
    border-radius: 999px;
    color: var(--muted-foreground);
    font-size: 0.72rem;
    font-weight: 700;
    padding: 0.12rem 0.45rem;
    white-space: nowrap;
  }

  .plan-table-wrap {
    border: 1px solid var(--border);
    border-radius: 8px;
    min-width: 0;
    overflow-x: auto;
  }

  .plan-table {
    border-collapse: collapse;
    min-width: 0;
    width: 100%;
  }

  .plan-table th,
  .plan-table td {
    border-top: 1px solid var(--border);
    font-size: 0.82rem;
    min-width: 7rem;
    overflow-wrap: anywhere;
    padding: 0.48rem 0.6rem;
    text-align: left;
    vertical-align: top;
  }

  .plan-table th {
    background: color-mix(in oklab, var(--muted) 32%, transparent);
    border-top: 0;
    color: var(--muted-foreground);
    font-size: 0.72rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .plan-table td {
    color: var(--muted-foreground);
  }

  .plan-table th:first-child,
  .plan-table td:first-child {
    min-width: 12rem;
  }

  .plan-table .primary-cell {
    color: var(--foreground);
    font-weight: 600;
  }

  .plan-pagination {
    justify-content: flex-end;
  }

  .plan-pagination span {
    color: var(--muted-foreground);
    font-size: 0.82rem;
    margin-right: auto;
  }

  .quiet-row {
    color: var(--muted-foreground);
    font-size: 0.88rem;
  }

  h3 {
    font-size: 1rem;
    margin: 0;
  }

  @media (max-width: 860px) {
    .plan-table th,
    .plan-table td {
      min-width: 10rem;
    }
  }
</style>
