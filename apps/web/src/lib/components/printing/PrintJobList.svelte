<script lang="ts">
import { t } from '$lib/presentation/localization';
import { timestampLabel } from '$lib/presentation/timestamp';
import { jobStatusLabel,reportedPrintOutcomeLabel } from '$lib/presentation/printing';
import type { ReportedPrintOutcome, PrintJob, RegisteredPrinter, PrintScope } from '$lib/domain/printing';
import * as Button from '$lib/components/ui/button/index.js';
import PrintJobResolution from './PrintJobResolution.svelte';
let { jobs, printers, scope, canPrint, busy = false, onCancel, onResolve }: {
    jobs: PrintJob[];
    printers: RegisteredPrinter[];
    scope: PrintScope;
    canPrint: boolean;
    busy?: boolean;
    onCancel: (job: PrintJob) => Promise<void>;
    onResolve?:(job:PrintJob,outcome:ReportedPrintOutcome)=>Promise<void>;
} = $props();
</script>
{#if jobs.length===0}<p>{t('web.Printing.emptyJobs')}</p>{:else}
 <ul class="job-list">
 {#each jobs as job(job.id)}
  {@const printer=printers.find(p=>p.id===job.printerId)}
  <li>
   <div class="job-heading"><strong>{printer?.name??job.printerId}</strong><span role="status">{job.resolution?t('web.Printing.userResolved'):jobStatusLabel(job.status)}</span></div>
   <p>{timestampLabel(job.createdAt)} · {t('web.Printing.completedCopies',{completed:job.completedCopies,copies:job.copies})}</p>
   {#if job.status==='queued'&&printer?.readiness!=='ready'}<p>{t('web.Printing.waitingPrinter')}</p>{/if}
   {#if job.status==='uncertain'}
    <p role="status">{t('web.Printing.uncertainHelp')}</p>
    {#if !job.idleConfirmedAt}<p>{t('web.Printing.waitingIdleConfirmation')}</p>
    {:else if canPrint&&onResolve}{#key job.revision}<PrintJobResolution {job} {busy} {onResolve}/>{/key}{/if}
   {/if}
   {#if job.resolution}<p>{t('web.Printing.reportedOutcome')}: {reportedPrintOutcomeLabel(job.resolution.reportedOutcome)}</p><p>{t('web.Printing.resolutionActorTime',{actor:job.resolution.resolvedBy,time:timestampLabel(job.resolution.resolvedAt)})}</p>{/if}
   {#if canPrint&&(job.status==='queued'||job.status==='claimed')}<Button.Root variant="outline" disabled={busy} onclick={()=>void onCancel(job)}>{t('web.Printing.cancelJob')}</Button.Root>{/if}
   {#if canPrint&&job.assetId&&['completed','failed','canceled'].includes(job.status)}<Button.Root variant="outline" href={`/tenants/${encodeURIComponent(scope.tenantId)}/inventories/${encodeURIComponent(scope.inventoryId)}/assets/${encodeURIComponent(job.assetId)}`}>{t('web.Printing.openAsset')}</Button.Root>{/if}
  </li>
 {/each}
 </ul>
{/if}
<style>.job-list{list-style:none;padding:0;display:grid;gap:var(--space-4)}li{border-bottom:1px solid var(--border);padding-block:var(--space-3);min-width:0}.job-heading{display:flex;flex-wrap:wrap;justify-content:space-between;gap:var(--space-2)}strong,p{overflow-wrap:anywhere}p{margin-block:var(--space-2);color:var(--muted-foreground)}</style>
