<script lang="ts">
import { t } from '$lib/presentation/localization';
import { timestampLabel } from '$lib/presentation/timestamp';
import { jobStatusLabel,reportedPrintOutcomeLabel } from '$lib/presentation/printing';
import type { ReportedPrintOutcome, PrintJob, RegisteredPrinter, PrintScope } from '$lib/domain/printing';
import * as Button from '$lib/components/ui/button/index.js';
import PrintJobResolution from './PrintJobResolution.svelte';
let { jobs, printers, scope, canPrint, busy = false, onCancel, onResolve, onReprint }: {
    jobs: PrintJob[];
    printers: RegisteredPrinter[];
    scope: PrintScope;
    canPrint: boolean;
    busy?: boolean;
    onCancel: (job: PrintJob) => Promise<void>;
    onReprint?: (job: PrintJob, trigger: HTMLElement) => void;
    onResolve?:(job:PrintJob,outcome:ReportedPrintOutcome)=>Promise<void>;
} = $props();
</script>
{#if jobs.length===0}<p>{t('web.Printing.emptyJobs')}</p>{:else}
 <ul class="job-list">
 {#each jobs as job(job.id)}
  {@const printer=printers.find(p=>p.id===job.printerId)}
  <li>
   <div class="job-heading"><strong>{printer?.name??t('web.Printing.removedPrinter')}</strong><span role="status">{job.resolution?t(job.resolution.reportedOutcome==='printed'?'web.Printing.personReportedPrinted':job.resolution.reportedOutcome==='not_printed'?'web.Printing.personReportedNotPrinted':'web.Printing.personReportedUnknown'):jobStatusLabel(job.status)}</span></div>
   <p>{timestampLabel(job.createdAt)}{#if !job.resolution} · {t('web.Printing.completedCopies',{completed:job.completedCopies,copies:job.copies})}{/if}</p>
   {#if job.status==='queued'&&printer?.readiness!=='ready'}<p>{t('web.Printing.waitingPrinter')}</p>{/if}
   {#if job.status==='uncertain'}
    <p role="status">{t('web.Printing.uncertainHelp')}</p>
    {#if !job.idleConfirmedAt}<p>{t('web.Printing.waitingIdleConfirmation')}</p>
    {:else if canPrint&&onResolve}{#key job.revision}<PrintJobResolution {job} {busy} {onResolve}/>{/key}{/if}
   {/if}
   {#if job.resolution}<p>{t('web.Printing.personReport')}</p><details><summary>{t('web.Printing.reportDetails')}</summary><p>{t('web.Printing.reportedOutcome')}: {reportedPrintOutcomeLabel(job.resolution.reportedOutcome)}</p><p>{t('web.Printing.completedCopies',{completed:job.completedCopies,copies:job.copies})}</p><p>{t('web.Printing.resolutionActorTime',{actor:job.resolution.resolvedBy,time:timestampLabel(job.resolution.resolvedAt)})}</p></details>{/if}
   {#if canPrint&&(job.status==='queued'||job.status==='claimed')}<Button.Root variant="outline" disabled={busy} onclick={()=>void onCancel(job)}>{t('web.Printing.cancelJob')}</Button.Root>{/if}
   {#if canPrint&&job.assetId&&['completed','failed','canceled'].includes(job.status)}{#if onReprint}<Button.Root variant="outline" disabled={busy} onclick={event=>onReprint?.(job,event.currentTarget)}>{t('web.Printing.printAgain')}</Button.Root>{:else}<Button.Root variant="outline" href={`/tenants/${encodeURIComponent(scope.tenantId)}/inventories/${encodeURIComponent(scope.inventoryId)}/assets/${encodeURIComponent(job.assetId)}`}>{t('web.Printing.openAsset')}</Button.Root>{/if}{/if}
  </li>
 {/each}
 </ul>
{/if}
<style>.job-list{list-style:none;padding:0;display:grid;gap:var(--space-4)}li{border-bottom:1px solid var(--border);padding-block:var(--space-3);min-width:0}.job-heading span{font-size:var(--text-metadata-size);font-weight:600;overflow-wrap:anywhere;background:var(--muted);border-radius:var(--radius-sm);padding:.25rem .5rem}.job-heading{display:flex;flex-wrap:wrap;justify-content:space-between;gap:var(--space-2)}strong,p{overflow-wrap:anywhere}p{margin-block:var(--space-2);color:var(--muted-foreground)}details{font-size:var(--text-metadata-size);min-width:0}summary{cursor:pointer;min-height:44px;align-content:center;color:var(--muted-foreground)}summary:focus-visible{outline:2px solid var(--ring);outline-offset:3px}li{overflow-wrap:anywhere}.job-heading strong{min-width:0}</style>
